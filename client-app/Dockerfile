# uses node.js to build the react app
FROM node:18-alpine as build

# sets the working directory inside the container
WORKDIR /app

# copies dependency files first (for better caching)
COPY package*.json ./

# installs the dependencies
RUN npm install

# copies the rest of the source code
COPY . .

# builds the static files (html/css/js)
RUN npm run build

# uses nginx to serve the files
FROM nginx:alpine

# copies the built files from the 'build' stage to nginx folder
COPY --from=build /app/dist /usr/share/nginx/html

# exposes port 80 for web traffic
EXPOSE 80

# starts the nginx server with global directives and keeps it running in the foreground
CMD ["nginx", "-g", "daemon off;"]