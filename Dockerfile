FROM golang:1.18-alpine

WORKDIR /matchingAppChatService

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN go build -o /matchingAppChatService

EXPOSE 8081

CMD [ "/matchingAppChatService" ]