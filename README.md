# Code Judge

A code judge platform with separated frontend and backend architecture.

## Project Structure

The project is divided into two main components:

### Backend
- Located in the `backend/` directory
- REST APIs for managing users, problems, and submissions
- Connected to PostgreSQL database
- Implemented with Go and Gin

### Frontend
- Located in the `frontend/` directory
- Web UI for interacting with the system
- Implemented with Go, Gin, and HTML/CSS/JavaScript

## Running the Project

### Requirements
- Docker and Docker Compose

### Run Commands
```bash
# Clone the repository
git clone https://github.com/yousefi-abolfazl/code-judge.git
cd code-judge

# Run with Docker
docker-compose up --build
```

### Service Access
- Frontend: http://localhost:2020
- Backend API: http://localhost:8080
- Database: PostgreSQL on port 5432

## Environment Variables
Each service has its own environment variables:

### Backend
Settings in `backend/config/config.yaml`:
- Database configuration
- JWT encryption key
- Application settings

### Frontend
Settings via environment variables:
- `API_URL`: Backend API address (default: http://backend:8080)