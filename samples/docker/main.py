#!/usr/bin/env python

from __future__ import absolute_import
from __future__ import division
from __future__ import print_function

import argparse

import tensorflow as tf

FLAGS = None

def train(fraction=1.0):
	config = tf.ConfigProto()
	config.gpu_options.per_process_gpu_memory_fraction = fraction

	a = tf.constant([1.0, 2.0, 3.0, 4.0, 5.0, 6.0], shape=[2, 3], name='a')
	b = tf.constant([1.0, 2.0, 3.0, 4.0, 5.0, 6.0], shape=[3, 2], name='b')
	c = tf.matmul(a, b)
    # Creates a session with log_device_placement set to True.
	config = tf.ConfigProto()
	config.gpu_options.per_process_gpu_memory_fraction = fraction
	sess = tf.Session(config=config)
	# Runs the op.
	while True:
		sess.run(c)


if __name__ == '__main__':
	train(0.7)
