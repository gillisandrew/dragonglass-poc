const esbuild = require('esbuild');
const { sassPlugin } = require('esbuild-sass-plugin');
const fs = require('fs');
const path = require('path');

async function build() {
  const isProd = process.env.NODE_ENV === 'production';
  
  console.log(`Building in ${isProd ? 'production' : 'development'} mode...`);

  // Ensure dist directory exists
  if (!fs.existsSync('dist')) {
    fs.mkdirSync('dist', { recursive: true });
  }

  // Build main.js
  await esbuild.build({
    entryPoints: ['src/main.ts'],
    bundle: true,
    outfile: 'dist/main.js',
    format: 'cjs',
    target: 'es2020',
    minify: isProd,
    sourcemap: !isProd,
    external: ['obsidian'],
  });

  // Build styles.css
  await esbuild.build({
    entryPoints: ['src/styles.scss'],
    bundle: true,
    outfile: 'dist/styles.css',
    minify: isProd,
    sourcemap: !isProd,
    plugins: [sassPlugin()],
  });

  // Copy manifest
  fs.copyFileSync('src/manifest.json', 'dist/manifest.json');
  
  console.log('✅ Build completed successfully!');
  console.log('📁 Output files:');
  console.log('  - dist/main.js');
  console.log('  - dist/styles.css');
  console.log('  - dist/manifest.json');
}

build().catch((error) => {
  console.error('❌ Build failed:', error);
  process.exit(1);
});