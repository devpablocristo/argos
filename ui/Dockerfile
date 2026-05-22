FROM node:20-alpine AS build

WORKDIR /app
ARG VITE_ARGOS_DEFAULT_SOURCE_URI=../sample
ENV VITE_ARGOS_DEFAULT_SOURCE_URI=$VITE_ARGOS_DEFAULT_SOURCE_URI
COPY package.json package-lock.json* ./
RUN npm install
COPY . .
RUN npm run build

FROM nginx:1.27-alpine
COPY --from=build /app/dist /usr/share/nginx/html
