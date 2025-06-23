const express = require('express');
const axios = require('axios');
const app = express();

app.use(express.json());

app.post('/chat', async (req, res) => {
  const { prompt } = req.body;

  try {
    const response = await axios.post('http://localhost:11434/api/generate', {
      model: 'llama3.2:1b',
      prompt,
      stream: false
    });

    res.json({ response: response.data.response });
  } catch (error) {
    console.error(error.message);
    res.status(500).json({ error: 'LLaMA server error' });
  }
});

app.get('/', (req, res) => {
  res.send('LLaMA.js Backend is running');
});

app.listen(3000, '0.0.0.0' , () => {
  console.log('Backend running on http://localhost:3000');
});
