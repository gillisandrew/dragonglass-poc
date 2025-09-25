import fs from 'fs';
import esbuild from 'esbuild';
import { sassPlugin } from 'esbuild-sass-plugin';
import { builtinModules } from 'module';
import manifestJson from './manifest.json' with { type: 'json' };


const isProd = process.env.NODE_ENV === 'production';

console.log(`Building in ${isProd ? 'production' : 'development'} mode...`);

// Ensure dist directory exists
if (!fs.existsSync('dist')) {
  fs.mkdirSync('dist', { recursive: true });
}

const banner = `/**
* This file is auto-generated. Do not edit directly.
* Generated on: ${new Date().toISOString()}
* Plugin Name: ${manifestJson.name}
* Plugin Version: ${manifestJson.version}
*/`;

// Build main.js
await esbuild.build({
  banner: {
    js: banner,
    css: banner,
  },
  entryPoints: ['src/main.ts', 'src/styles.scss'],
  bundle: true,
  outdir: 'dist',
  format: 'cjs',
  target: 'es2020',
  minify: false,
  sourcemap: false,
  external: [
    "obsidian",
    "electron",
    "@codemirror/autocomplete",
    "@codemirror/closebrackets",
    "@codemirror/collab",
    "@codemirror/commands",
    "@codemirror/comment",
    "@codemirror/fold",
    "@codemirror/gutter",
    "@codemirror/highlight",
    "@codemirror/history",
    "@codemirror/language",
    "@codemirror/lint",
    "@codemirror/matchbrackets",
    "@codemirror/panel",
    "@codemirror/rangeset",
    "@codemirror/rectangular-selection",
    "@codemirror/search",
    "@codemirror/state",
    "@codemirror/stream-parser",
    "@codemirror/text",
    "@codemirror/tooltip",
    "@codemirror/view",
    ...builtinModules,
  ],
  plugins: [sassPlugin()],
});

// Create manifest
fs.writeFileSync('dist/manifest.json', JSON.stringify(manifestJson, null, 2));

console.log('✅ Build completed successfully!');
console.log('📁 Output files:');
console.log('  - dist/main.js');
console.log('  - dist/styles.css');
console.log('  - dist/manifest.json');