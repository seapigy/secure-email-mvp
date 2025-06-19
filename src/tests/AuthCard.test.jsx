import { render, screen, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import AuthCard from '../components/AuthCard';
import * as api from '../lib/api';
jest.mock('../lib/api');

test('renders login form', () => {
  render(<AuthCard />, { wrapper: MemoryRouter });
  expect(screen.getByLabelText(/email input/i)).toBeInTheDocument();
  expect(screen.getByLabelText(/password input/i)).toBeInTheDocument();
  expect(screen.getByLabelText(/totp code input/i)).toBeInTheDocument();
});

test('toggles to sign-up form', () => {
  render(<AuthCard />, { wrapper: MemoryRouter });
  fireEvent.click(screen.getByText(/sign up/i));
  expect(screen.getByLabelText(/confirm password input/i)).toBeInTheDocument();
});

test('shows error toast on invalid sign-up', async () => {
  render(<AuthCard />, { wrapper: MemoryRouter });
  fireEvent.click(screen.getByText(/sign up/i));
  fireEvent.change(screen.getByLabelText(/email input/i), { target: { value: 'invalid@example.com' } });
  fireEvent.change(screen.getByLabelText(/password input/i), { target: { value: 'password123' } });
  fireEvent.change(screen.getByLabelText(/confirm password input/i), { target: { value: 'password123' } });
  fireEvent.click(screen.getByText(/sign up/i));
  expect(await screen.findByText(/invalid email format/i)).toBeInTheDocument();
});

test('submits sign-up and shows QR code', async () => {
  api.signup.mockResolvedValue({ data: { temp_id: 'uuid', totp_qr: 'base64_png' } });
  render(<AuthCard />, { wrapper: MemoryRouter });
  fireEvent.click(screen.getByText(/sign up/i));
  fireEvent.change(screen.getByLabelText(/email input/i), { target: { value: 'test@securesystem.email' } });
  fireEvent.change(screen.getByLabelText(/password input/i), { target: { value: 'password123' } });
  fireEvent.change(screen.getByLabelText(/confirm password input/i), { target: { value: 'password123' } });
  fireEvent.click(screen.getByText(/sign up/i));
  expect(api.signup).toHaveBeenCalledWith({
    email: 'test@securesystem.email',
    password: 'password123',
    confirm_password: 'password123',
  });
  expect(await screen.findByText(/scan this QR code/i)).toBeInTheDocument();
}); 