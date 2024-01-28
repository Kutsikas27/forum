## docker

docker build -t forum .
docker run --name forum_c1 -p 8080:8080 -d forum
docker start forum_c1