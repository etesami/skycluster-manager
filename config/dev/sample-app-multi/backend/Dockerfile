# Use the official Python image as the base image
FROM python:3.6-slim

# Set the working directory in the container
WORKDIR /app

# Copy the requirements file and install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy the application code
COPY backend.py .

# Expose the port the app runs on
EXPOSE 5000

# Start the application
CMD ["python", "backend.py"]