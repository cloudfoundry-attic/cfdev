#!/bin/bash

# Director IP
sudo ifconfig lo0 inet 10.245.0.2/32 remove

# CF Router IP
sudo ifconfig lo0 inet 10.244.0.34/32 remove