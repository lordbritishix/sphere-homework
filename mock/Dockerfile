# Use the Node.js official image as a base
FROM node:20

# Set the working directory inside the container
WORKDIR /usr/src/app

# Copy the rest of the application files
COPY . .

# Specify the default command to run the application
ENTRYPOINT ["node", "mockFxRateSender.js"]

CMD []