FROM node:18

WORKDIR /app

COPY . .

RUN npm install
RUN npx tsc

EXPOSE 3000

CMD [ "node", "./dist/app.js" ]