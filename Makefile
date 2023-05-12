#docker_repository ?= 594471699039.dkr.ecr.us-east-1.amazonaws.com/auto-heapdump
docker_repository ?= github.com/lowrielabs
git_commit_sha ?= $(shell git rev-parse --short HEAD)
oom_repo_suffix = /oom
monitor_repo_suffix = /monitor

ci: docker

all:	ci

docker:	docker_build docker_push

docker_build:
	#( env | grep -E -q '^AWS_ACCESS_KEY_ID=[A-Za-z0-9]' ) || (echo "ERROR: Environment variable 'AWS_ACCESS_KEY_ID' is not set" 1>&2 && false )
	#( env | grep -E -q '^AWS_SECRET_ACCESS_KEY=[A-Za-z0-9]' ) || (echo "ERROR: Environment variable 'AWS_SECRET_ACCESS_KEY' is not set" 1>&2 && false )
	#( env | grep -E -q '^AWS_S3_BUCKET=[A-Za-z0-9]' ) || (echo "ERROR: Environment variable 'AWS_S3_BUCKET' is not set" 1>&2 && false )
	docker build -f Dockerfile-oom \
		-t $(docker_repository)$(oom_repo_suffix):latest .
	docker build -f Dockerfile-monitor \
		-t $(docker_repository)$(monitor_repo_suffix):latest \
		--build-arg AWS_ACCESS_KEY_ID=$(AWS_ACCESS_KEY_ID) \
		--build-arg AWS_SECRET_ACCESS_KEY=$(AWS_SECRET_ACCESS_KEY) \
		--build-arg AWS_S3_BUCKET=$(AWS_S3_BUCKET) \
		.

docker_push:
	#docker push $(docker_repository)$(oom_repo_suffix):$(git_commit_sha)
	#docker push $(docker_repository)$(oom_repo_suffix):latest
	#docker push $(docker_repository)$(monitor_repo_suffix):$(git_commit_sha)
	#docker push $(docker_repository)$(monitor_repo_suffix):latest

go:
	cd hoover && env GOOS=linux GOARCH=amd64 go build -o main .

java:
	javac -source 11 -target 11 -cp `ls lib/*jar | sed 's/ /:/g' `:src/main/java src/main/java/com/jumbly/monitoring/Main.java
	cd src/main/java && jar cmf META-INF/MANIFEST.MF oom.jar `find . -name \*.class`

clean:
	find src/main/java -name "*.class" | xargs rm || true
	rm -f src/main/java/oom.jar || true
	rm -f hoover/main || true

deploy:
	# These commands are for minikube only, and purely to demonstrate the functionality
	# of auto-generating (and grabbing) heapdumps
	eval $(minikube docker-env) 
	kubectl delete pod heapy-demo || true
	kubectl apply -f heapy-demo.yml

