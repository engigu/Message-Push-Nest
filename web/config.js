// config.js

const isProduction = process.env.NODE_ENV === 'prod';

const config = {
    apiUrl: isProduction ? '' : 'http://localhost:8000',
};

export default config;

