
# WebSocket Application

This application allows multiple clients to connect via WebSocket and send messages to each other. Below are the instructions to run the application.

## Steps to Run the Application

### 1. Clone the Repository

Clone the repository to your local machine by running the following command:

```bash
git clone https://github.com/AryanGKulkarni/OptionX_Assignment
cd OptionX_Assignment
```

### 2. Install Packages

Clone the repository to your local machine by running the following command:
```
go get github.com/gorilla/websocket 
go get github.com/google/uuid
```
### 3. Run the Application
To start the application, run the following command in the root directory:
```
go run main.go
```
### 4. Open WebSocket Connection in Postman
Once the application is running, open Postman and establish a WebSocket connection using the following URL:
```
ws://localhost:8080/ws
```
### 5. Test with Multiple Clients
To simulate multiple clients, open multiple tabs in Postman and connect to the same WebSocket URL.

### 6. Receive the Welcome Message
After being connected, you will receive a message in the following format:
```
Welcome! Your Client ID: 2ef1e79b-11a2-4da6-a6a5-a78b2ed45b36
Connected Clients:
7b83069c-90b7-41cd-be00-50c8ec195caf
f38a1c85-bb92-4fd3-bcd3-288fe758c58d
2ef1e79b-11a2-4da6-a6a5-a78b2ed45b36 (You)
```
### 7. Select a Client and Send a Message
To send a message to another client, select the client you want to send the message to and send the following message format:
```
{
    "id": "f38a1c85-bb92-4fd3-bcd3-288fe758c58d",
    "message": "Hello!"
}
```
### 8. Demo Video
You can watch a demo video of the application here:

[Demo Video](https://drive.google.com/file/d/1ODg3pVChGxu_JoRzda3iHDBA3siqq8XO/view?usp=sharing)

