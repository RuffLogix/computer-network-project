# Frontend Dockerfile
FROM node:18-alpine

# Install pnpm
RUN npm install -g pnpm

WORKDIR /app

# Copy package files
COPY package.json pnpm-lock.yaml ./

# Install dependencies with pnpm (including dev dependencies for build)
RUN pnpm install --frozen-lockfile && pnpm rebuild lightningcss

# Copy source code
COPY . .

# Build the app
RUN pnpm build

# Expose port
EXPOSE 3000

# Start the app
CMD ["pnpm", "start"]