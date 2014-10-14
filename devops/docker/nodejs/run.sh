#!/bin/bash
echo "Start Supervisor"
exec supervisord -e debug -n
