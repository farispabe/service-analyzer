# Stage 1: Build React app
FROM node:18 AS build

WORKDIR /app

# Set environment variable to fix OpenSSL error
ENV NODE_OPTIONS=--openssl-legacy-provider

# Install dependencies
COPY package.json ./
RUN npm install

# Copy all files and build the React app
COPY . .
RUN npm run build

# Stage 2: Serve the React app using Nginx
FROM nginx:alpine
COPY --from=build /app/build /usr/share/nginx/html
EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
