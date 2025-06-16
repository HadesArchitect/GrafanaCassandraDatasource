// Jest setup provided by Grafana scaffolding
import './.config/jest-setup';

// Additional setup for this project
import React from 'react';

// Polyfill for React.useId (React 18 feature) for React 17 compatibility
if (!React.useId) {
  let idCounter = 0;
  React.useId = () => {
    return `:r${(idCounter++).toString(36)}:`;
  };
}
