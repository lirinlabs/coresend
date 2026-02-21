const fs = require('fs');
const path = require('path');

const indexPath = path.join(__dirname, 'dist/index.html');

let html = fs.readFileSync(indexPath, 'utf-8');

// Add nonce placeholder to script tags - flexible regex that matches regardless of attribute order
html = html.replace(
    /<script\b([^>]*)\btype="module"([^>]*)>/g,
    (match, before, after) =>
        `<script${before}type="module" nonce="__CSP_NONCE__"${after}>`,
);

// Verify replacement occurred
if (!html.includes('nonce="__CSP_NONCE__"')) {
    console.error(
        'ERROR: Failed to inject CSP nonce placeholder. The regex did not match any script tags.',
    );
    process.exit(1);
}

fs.writeFileSync(indexPath, html);

console.log('CSP nonce placeholder injected into dist/index.html');
