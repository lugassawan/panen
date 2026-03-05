function createCommandPaletteStore() {
  let open = $state(false);

  return {
    get open() {
      return open;
    },
    toggle() {
      open = !open;
    },
    close() {
      open = false;
    },
  };
}

export const commandPalette = createCommandPaletteStore();
