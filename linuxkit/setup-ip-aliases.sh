#!/bin/bash

# Director IP
ip addr add 10.245.0.2/32 dev lo0

# CF Router IP
ip addr add 10.244.0.34/32 dev lo0