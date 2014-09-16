#!/bin/bash
cd /home/app/
npm install --production
bower install --config.interactive=false --allow-root
PORT=3001 npm start
