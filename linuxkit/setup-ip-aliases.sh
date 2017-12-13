#!/bin/bash

# Director IP
ifconfig lo0 add 10.245.0.2/32

# CF Router IP
ifconfig lo0 add 10.244.0.34/32