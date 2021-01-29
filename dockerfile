From 'golang'
Run mkdir -p /app
ADD ./main /app/main
ADD ./config.json /app/config.json
WORKDIR /app
ENTRYPOINT './main'