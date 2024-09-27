# aws-vide-oh
Master's thesis project: Serverless AWS video streaming app 

# Features
The system recognizes 3 basic types of users:
- Unregistered user
- Registered user
- Administrator

Below are the functionalities that each type of user has access to.

Unregistered user:
- Registration
- Login
- Search for existing videos (with thumbnails that are automatically extracted when videos are uploaded (courtesy of _FFmpeg_))
- Watch videos (the videos will be streamed, not downloaded, and viewed within this application), control playback speed
- Download videos

Registered user:

- Search for existing videos
- Watch videos
- Download videos
- Upload videos to the system (along with details such as title, description, etc.)
- View, edit, and delete their own videos
- View the average rating and all comments on a video
- CRUD (Create, Read, Update, Delete) their own comments and ratings on each video
- Edit profile info

Administrator:

- All the functionalities that a registered user has access to (with the exception of being unable to upload videos)
- User blocking (with notification via email)
- Deletion of videos or modification of video data deemed inappropriate
- Deletion of inappropriate comments

# Architecture
![AWS arhitektura drawio (2)](https://github.com/user-attachments/assets/2b32bd6f-b63a-4ea0-9722-d50ef8fab789)

