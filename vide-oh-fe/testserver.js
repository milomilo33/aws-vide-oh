// Optional: allow environment to specify port
const port = process.env.PORT || 8080;

// Import required modules
const express = require('express');
const path = require('path');

// Create server instance
const app = express();

app.use((req, res, next) => {
    console.log(`Request URL: ${req.url}`);
    next();
});

// Serve static files from the 'dist' directory
app.use(express.static('dist'));

// Add a wildcard route to serve index.html for any unknown paths
app.get('*', (req, res) => {
    res.sendFile(path.resolve(__dirname, 'dist', 'index.html'));
});

// Start the server
app.listen(port, () => console.log(`Listening on port ${port}`));
