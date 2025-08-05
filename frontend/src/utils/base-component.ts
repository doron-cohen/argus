import { type ReadableAtom } from "nanostores";

export abstract class BaseComponent extends HTMLElement {
  private subscriptions: (() => void)[] = [];

  // Automatic cleanup on disconnect
  disconnectedCallback() {
    this.subscriptions.forEach((unsub) => unsub());
    this.subscriptions = [];
  }

  // Helper for state binding
  protected bindState<T>(store: ReadableAtom<T>, handler: (value: T) => void) {
    const unsub = store.subscribe(handler);
    this.subscriptions.push(unsub);
  }
}
