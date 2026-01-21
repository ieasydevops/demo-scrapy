FROM node:18-alpine AS builder

WORKDIR /app

COPY package*.json ./
RUN npm ci --silent --registry=https://registry.npmmirror.com || npm ci --silent

COPY . .
RUN npm run build

FROM nginx:alpine

RUN apk add --no-cache wget

COPY --from=builder /app/dist /usr/share/nginx/html

COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
