import { create } from 'zustand';

/**
 * Session State Interface
 * 
 * Manages session-specific state for secure email access.
 * Tracks which emails have been unlocked during the current session
 * and provides methods to manage this state.
 */
interface SessionState {
  /** Set of email IDs that have been unlocked in this session */
  unlockedEmails: Set<string>;
  
  /** Unlock an email by adding its ID to the unlocked set */
  unlockEmail: (emailId: string) => void;
  
  /** Check if an email has been unlocked in this session */
  isEmailUnlocked: (emailId: string) => boolean;
  
  /** Clear all unlocked emails (useful for logout or session reset) */
  clearUnlockedEmails: () => void;
}

/**
 * Session Store
 * 
 * Zustand store for managing session-specific state.
 * This store tracks which secure emails have been unlocked
 * during the current user session, supporting both per-session
 * and per-email password protection modes.
 * 
 * Features:
 * - Track unlocked emails across the session
 * - Support for per-email password protection
 * - Session-based email unlocking
 * - Clean state management for logout
 */
export const useSessionStore = create<SessionState>((set, get) => ({
  // Initialize with empty set of unlocked emails
  unlockedEmails: new Set(),
  
  /**
   * Unlock an email by adding its ID to the unlocked set
   * @param emailId - The unique identifier of the email to unlock
   */
  unlockEmail: (emailId: string) => {
    set((state) => ({
      unlockedEmails: new Set([...state.unlockedEmails, emailId])
    }));
  },
  
  /**
   * Check if an email has been unlocked in this session
   * @param emailId - The unique identifier of the email to check
   * @returns true if the email has been unlocked, false otherwise
   */
  isEmailUnlocked: (emailId: string) => {
    return get().unlockedEmails.has(emailId);
  },
  
  /**
   * Clear all unlocked emails from the session
   * This is typically called on logout or session reset
   */
  clearUnlockedEmails: () => {
    set({ unlockedEmails: new Set() });
  }
}));
