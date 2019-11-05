NAME = rad
IMAGE_PREFIX = joshrwolf
IMAGE_VERSION = latest

app:
	go build -v -o $(NAME) main.go

docker:
	DOCKER_BUILDKIT=1 docker build . -t $(IMAGE_PREFIX)/$(NAME):$(IMAGE_VERSION)