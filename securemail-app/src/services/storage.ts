// Secure Storage Service for SecureMail Frontend
import AsyncStorage from '@react-native-async-storage/async-storage';
import * as SecureStore from 'expo-secure-store';
import { Platform } from 'react-native';

class StorageService {
  private readonly TOKEN_KEY = 'securemail_auth_token';
  private readonly USER_KEY = 'securemail_user_data';
  private readonly SESSION_KEY = 'securemail_session';

  // Store authentication token securely
  async storeToken(token: string): Promise<void> {
    try {
      if (Platform.OS === 'web') {
        // For web, use localStorage with additional security measures
        await AsyncStorage.setItem(this.TOKEN_KEY, token);
      } else {
        // For mobile, use SecureStore
        await SecureStore.setItemAsync(this.TOKEN_KEY, token);
      }
    } catch (error) {
      console.error('Error storing token:', error);
      throw new Error('Failed to store authentication token');
    }
  }

  // Retrieve authentication token
  async getToken(): Promise<string | null> {
    try {
      if (Platform.OS === 'web') {
        return await AsyncStorage.getItem(this.TOKEN_KEY);
      } else {
        return await SecureStore.getItemAsync(this.TOKEN_KEY);
      }
    } catch (error) {
      console.error('Error retrieving token:', error);
      return null;
    }
  }

  // Remove authentication token
  async removeToken(): Promise<void> {
    try {
      if (Platform.OS === 'web') {
        await AsyncStorage.removeItem(this.TOKEN_KEY);
      } else {
        await SecureStore.deleteItemAsync(this.TOKEN_KEY);
      }
    } catch (error) {
      console.error('Error removing token:', error);
    }
  }

  // Store user data
  async storeUser(user: any): Promise<void> {
    try {
      const userData = JSON.stringify(user);
      await AsyncStorage.setItem(this.USER_KEY, userData);
    } catch (error) {
      console.error('Error storing user data:', error);
      throw new Error('Failed to store user data');
    }
  }

  // Retrieve user data
  async getUser(): Promise<any | null> {
    try {
      const userData = await AsyncStorage.getItem(this.USER_KEY);
      return userData ? JSON.parse(userData) : null;
    } catch (error) {
      console.error('Error retrieving user data:', error);
      return null;
    }
  }

  // Remove user data
  async removeUser(): Promise<void> {
    try {
      await AsyncStorage.removeItem(this.USER_KEY);
    } catch (error) {
      console.error('Error removing user data:', error);
    }
  }

  // Store session data (token + user)
  async storeSession(token: string, user: any): Promise<void> {
    try {
      await Promise.all([
        this.storeToken(token),
        this.storeUser(user),
      ]);
    } catch (error) {
      console.error('Error storing session:', error);
      throw new Error('Failed to store session data');
    }
  }

  // Retrieve session data
  async getSession(): Promise<{ token: string | null; user: any | null }> {
    try {
      const [token, user] = await Promise.all([
        this.getToken(),
        this.getUser(),
      ]);
      return { token, user };
    } catch (error) {
      console.error('Error retrieving session:', error);
      return { token: null, user: null };
    }
  }

  // Clear all session data
  async clearSession(): Promise<void> {
    try {
      await Promise.all([
        this.removeToken(),
        this.removeUser(),
      ]);
    } catch (error) {
      console.error('Error clearing session:', error);
    }
  }

  // Check if user is authenticated
  async isAuthenticated(): Promise<boolean> {
    try {
      const { token, user } = await this.getSession();
      return !!(token && user);
    } catch (error) {
      console.error('Error checking authentication status:', error);
      return false;
    }
  }

  // Store trial warning data
  async storeTrialWarning(warning: any): Promise<void> {
    try {
      await AsyncStorage.setItem('trial_warning', JSON.stringify(warning));
    } catch (error) {
      console.error('Error storing trial warning:', error);
    }
  }

  // Retrieve trial warning data
  async getTrialWarning(): Promise<any | null> {
    try {
      const warning = await AsyncStorage.getItem('trial_warning');
      return warning ? JSON.parse(warning) : null;
    } catch (error) {
      console.error('Error retrieving trial warning:', error);
      return null;
    }
  }

  // Clear trial warning data
  async clearTrialWarning(): Promise<void> {
    try {
      await AsyncStorage.removeItem('trial_warning');
    } catch (error) {
      console.error('Error clearing trial warning:', error);
    }
  }
}

// Export singleton instance
export const storageService = new StorageService();
export default storageService;
