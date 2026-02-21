const fs = require('fs');
const path = require('path');

const indexPath = path.join(__dirname, 'dist/index.html');

let html = fs.readFileSync(indexPath, 'utf-8');

// Add nonce placeholder to script tags
html = html.replace(
    /<script type="module" crossorigin/g,
    '<script type="module" nonce="__CSP_NONCE__" crossorigin',
);

fs.writeFileSync(indexPath, html);

console.log('CSP nonce placeholder injected into dist/index.html');
