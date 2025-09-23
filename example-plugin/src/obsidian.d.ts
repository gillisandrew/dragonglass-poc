// Simple type declarations for the example
declare module 'obsidian' {
  export class Plugin {
    addRibbonIcon(icon: string, title: string, callback: (evt: MouseEvent) => void): HTMLElement;
    addCommand(command: { id: string; name: string; callback: () => void }): void;
    onload(): Promise<void> | void;
    onunload(): void;
  }
  
  export class Notice {
    constructor(message: string);
  }
}