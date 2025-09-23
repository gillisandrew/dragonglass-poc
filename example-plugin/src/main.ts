import { Plugin, Notice } from 'obsidian';

export default class ExamplePlugin extends Plugin {
  async onload() {
    console.log('Loading Example Plugin');

    // Add ribbon icon
    this.addRibbonIcon('dice', 'Example Plugin', (evt: MouseEvent) => {
      new Notice('Example Plugin activated!');
    });

    // Add command
    this.addCommand({
      id: 'example-command',
      name: 'Example Command',
      callback: () => {
        new Notice('Example command executed!');
      }
    });
  }

  onunload() {
    console.log('Unloading Example Plugin');
  }
}