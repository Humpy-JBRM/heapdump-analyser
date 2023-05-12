FROM hoover:1.20-alpine as builder

WORKDIR /app

RUN go mod download
COPY hoover/ ./
RUN go mod tidy
RUN go build -o main .

FROM ubuntu
RUN apt update
ARG DEBIAN_FRONTEND=noninteractive
RUN apt install -y awscli inotify-tools unzip curl ca-certificates openjdk-11-jdk
COPY --from=builder /app/main /
RUN chmod 755 /main

# Eclipse MAT
RUN curl -L -o /var/tmp/mat.zip 'https://www.eclipse.org/downloads/download.php?file=/mat/1.14.0/rcp/MemoryAnalyzer-1.14.0.20230315-linux.gtk.x86_64.zip&r=1'
RUN unzip /var/tmp/mat.zip
RUN rm -f /var/tmp/mat.zip

USER nobody
RUN mkdir -p /heapdumps
RUN chmod 777 /heapdumps
ENV HEAPDUMPS_DIR "/heapdumps"
ENV HEAPDUMP_SERVICE_LISTEN "0.0.0.0:8000"

CMD ["/main", "-analyse=true", "-heapdumpsDir=${HEAPDUMPS_DIR}", "-listen=${HEAPDUMP_SERVICE_LISTEN}"]
