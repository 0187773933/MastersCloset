#!/bin/bash
APP_NAME="public-mct-server"
id=$(sudo docker restart $APP_NAME)
sudo docker logs -f $id