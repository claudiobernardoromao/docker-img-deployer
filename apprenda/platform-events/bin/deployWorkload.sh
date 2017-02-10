#!/bin/bash

# Modify docker-image executable perms
chmod 755 bin/docker-image
bin/docker-image deploy
