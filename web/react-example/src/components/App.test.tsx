import React from 'react';
import { render  } from '@testing-library/react';
import App from './App';

test('stub test', () => {
  render(<App />);
  expect(1).toEqual(1);
});
