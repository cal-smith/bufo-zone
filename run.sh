#!/usr/bin/env bash

python manage.py migrate
supervisord -c supervisor.conf