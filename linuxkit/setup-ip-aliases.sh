#!/bin/bash

# Director IP
sudo ifconfig lo0 add 10.245.0.2/32

# CF Router IP
sudo ifconfig lo0 add 10.144.0.34/32
