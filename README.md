# Health Trackers

A Web Application for tracking health metrics. You can sign up, create your own
health tracker, list associated symptoms, log events, and view your health history.

I've included AI to analyze your logs, and provide insights on your health from
a Naturopathic or Ayuvedic POV.

## Getting Started

### Requirements

For full-stack development, you will need [Node](http://nodejs.org/) and [Go](https://golang.org/)
installed in your environment.

### Install

    git clone https://github.com/ArvoyaDev/health-trackers-backend
    cd PROJECT

### Configure app

Create a `.env` file in the root directory and add the following:

```bash
AWS_DATABASE_URL=YOUR_AWS_DATABASE_URL
PORT=8080
OPENAI_API=YOUR_OPENAI_KEY
AWS_REGION=YOUR_AWS_REGION
DATABASE_NAME=YOUR_DATABASE_NAME
DATABASE_USER=YOUR_DATABASE_USER
RDS_ENDPOINT=YOUR_RDS_ENDPOINT
COGNITO_USER_POOL_ID=YOUR_COGNITO_USER_POOL_ID
COGNITO_APP_CLIENT_ID=YOUR_COGNITO_APP_CLIENT_ID
COGNITO_CLIENT_SECRET=YOUR_COGNITO_CLIENT_SECRET
AWS_ACCESS_KEY=YOUR_AWS_ACCESS_KEY
AWS_TOKEN_SIGNING_KEY=YOUR_AWS_TOKEN_SIGNING_KEY
ENV=dev
CA_CERT=YOUR_CA_CERT
```

### Start & watch

    go run .

## Frontend Architecture

### Languages & tools

- [Go](https://golang.org/)
- [AWS - RDS](https://aws.amazon.com/rds/)
- [AWS - Cognito](https://aws.amazon.com/cognito/)

## Change Log

### 0.0.1

## Collaborators

Self Developed Project

Open to collaboration.
