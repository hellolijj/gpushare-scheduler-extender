package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/gpushare"
	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/routes"
	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/scheduler"
	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/utils/signals"
	"github.com/AliyunContainerService/gpushare-scheduler-extender/pkg/policy"
	
	"github.com/comail/colog"
	"github.com/julienschmidt/httprouter"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const RecommendedKubeConfigPathEnv = "KUBECONFIG"

var (
	clientset    *kubernetes.Clientset
	resyncPeriod = 30 * time.Second
	clientConfig clientcmd.ClientConfig
)

func initKubeClient() {
	kubeConfig := ""
	if len(os.Getenv(RecommendedKubeConfigPathEnv)) > 0 {
		// use the current context in kubeconfig
		// This is very useful for running locally.
		kubeConfig = os.Getenv(RecommendedKubeConfigPathEnv)
	}

	// Get kubernetes config.
	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		log.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	// create the clientset
	clientset, err = kubernetes.NewForConfig(restConfig)
	if err != nil {
		log.Fatalf("fatal: Failed to init rest config due to %v", err)
	}
}

func main() {
	// Call Parse() to avoid noisy logs
	flag.CommandLine.Parse([]string{})

	colog.SetDefaultLevel(colog.LInfo)
	colog.SetMinLevel(colog.LInfo)
	colog.SetFormatter(&colog.StdFormatter{
		Colors: true,
		Flag:   log.Ldate | log.Ltime | log.Lshortfile,
	})
	colog.Register()
	level := StringToLevel(os.Getenv("LOG_LEVEL"))
	log.Print("Log level was set to ", strings.ToUpper(level.String()))
	colog.SetMinLevel(level)
	threadness := StringToInt(os.Getenv("THREADNESS"))

	initKubeClient()
	port := os.Getenv("PORT")
	if _, err := strconv.Atoi(port); err != nil {
		port = "39997"
	}

	var schedulerPolicy, staticDgx string
	flag.StringVar(&schedulerPolicy, "policy", "", "config gpu select policy, , for more detail: https://github.com/hellolijj/")
	flag.StringVar(&staticDgx, "staticdgx", "", "config static node dgx, for more detail: https://github.com/hellolijj/")
	flag.Parse()
	if len(schedulerPolicy) == 0 {
		schedulerPolicy = "simple"
	}
	if schedulerPolicy != "simple" && schedulerPolicy != "best_effort" && schedulerPolicy != "static" {
		log.Printf("uninvalid gpu policy %v", schedulerPolicy)
		return
	}
	// to valid static gpu config

	// Set up signals so we handle the first shutdown signal gracefully.
	stopCh := signals.SetupSignalHandler()

	informerFactory := kubeinformers.NewSharedInformerFactory(clientset, resyncPeriod)
	controller, err := gpushare.NewController(clientset, informerFactory, stopCh)
	if err != nil {
		log.Fatalf("Failed to start due to %v", err)
	}
	err = controller.BuildCache()
	if err != nil {
		log.Fatalf("Failed to start due to %v", err)
	}

	go controller.Run(threadness, stopCh)

	policy, err := policy.NewPolicy(schedulerPolicy, staticDgx)
	if err != nil {
		log.Fatalf("Failed to build policy due to %v", err)
	}

	gpuTopologyPrioritize := scheduler.NewGPUTopologyPrioritize(clientset, controller.GetSchedulerCache(), policy)
	gpuTopologyBind := scheduler.NewGPUShareBind(clientset, controller.GetSchedulerCache(), policy)
	gpuTopologyInspect := scheduler.NewGPUTopologyInspect(controller.GetSchedulerCache())

	router := httprouter.New()

	routes.AddPProf(router)
	routes.AddVersion(router)
	routes.AddPrioritize(router, gpuTopologyPrioritize)
	routes.AddBind(router, gpuTopologyBind)
	routes.AddInspect(router, gpuTopologyInspect)

	log.Printf("info: server starting on the port :%s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal(err)
	}
}

func StringToLevel(levelStr string) colog.Level {
	switch level := strings.ToUpper(levelStr); level {
	case "TRACE":
		return colog.LTrace
	case "DEBUG":
		return colog.LDebug
	case "INFO":
		return colog.LInfo
	case "WARNING":
		return colog.LWarning
	case "ERROR":
		return colog.LError
	case "ALERT":
		return colog.LAlert
	default:
		log.Printf("warning: LOG_LEVEL=\"%s\" is empty or invalid, fallling back to \"INFO\".\n", level)
		return colog.LInfo
	}
}

func StringToInt(sThread string) int {
	thread := 1

	return thread
}
