FROM golang:1.16-alpine
#FROM arm64v8/golang:1.19-alpine
ENV GOOS "linux"
ENV GOARCH "arm64"
WORKDIR /app

COPY *.go ./
COPY go.mod ./
RUN go get -d -v
RUN go build -o ./go-microservice

FROM docker.io/jctarla/oci-kubectl:arm64

WORKDIR /app
EXPOSE 8080
COPY entrypoint.sh ./
#RUN chmod +x entrypoint.sh 
COPY --from=0 /app/go-microservice ./
CMD [ "./entrypoint.sh" ]



