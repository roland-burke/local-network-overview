#Build container
docker build . -t local-network-overview

# Start container
docker run -p 127.0.0.1:8081:8080 -v ./conf/config.json:/conf/config.json local-network-overview