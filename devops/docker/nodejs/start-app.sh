#!/bin/bash
cd /home/app/
npm install --production
bower install --allow-root
PORT=3001 NODE_ENV=production npm start
