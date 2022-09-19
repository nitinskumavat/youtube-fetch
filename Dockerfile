FROM golang:1.18


WORKDIR /app
COPY go.mod .
COPY go.sum .


RUN go mod download
COPY . .

RUN go build

ENV MONGODB_URI="mongodb+srv://nitin:nitin123.@videocluster.8coh86y.mongodb.net/?retryWrites=true&w=majority"
CMD [ "./youtube-fetch" ]