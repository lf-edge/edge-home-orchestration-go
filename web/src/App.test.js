import { render, screen } from '@testing-library/react';
import App from './App';

test('renders homeedge link', () => {
  render(<App />);
  const linkElement = screen.getByAltText(/logo/i);
  expect(linkElement).toBeInTheDocument();
});
