#!/usr/bin/env sh
set -eu
envsubst '${BACKEND_URL}' < /etc/nginx/nginx.conf.tmpl > /etc/nginx/nginx.conf
exec nginx -g "daemon off;"
