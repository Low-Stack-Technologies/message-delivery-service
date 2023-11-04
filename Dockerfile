FROM node:21-alpine

WORKDIR /app

VOLUME [ "/data" ]

ENV DATA_PATH=/data
ENV CONFIGURATION_FILE_PATH=/data/config.json

RUN npm install -g pnpm

COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

COPY . .

CMD [ "pnpm", "tsx", "./src/app.ts" ]