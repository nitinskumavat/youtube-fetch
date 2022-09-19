FROM golang:1.18


WORKDIR /app
COPY go.mod .
COPY go.sum .


RUN go mod download
COPY . .

RUN go build

ENV MONGODB_URI="mongodb+srv://nitin:nitin123.@videocluster.8coh86y.mongodb.net/?retryWrites=true&w=majority"
ENV API_KEYS="AIzaSyDgrxbHVSHMVN3Yt8X7rucMtArmAeWunR8#AIzaSyC2iveqSEY9W3JpMfINdyiR181foX7cYaw#AIzaSyBPlrqi0X1vdf57cXvy4mxUVbCPt1D2t1w"
CMD [ "./youtube-fetch" ]