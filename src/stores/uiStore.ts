import { create } from 'zustand';
import { persist } from 'zustand/middleware';
import type { UIState, EmailFilters } from '@/types';

interface UIStore extends UIState {
  // Actions
  setTheme: (theme: 'light' | 'dark' | 'system') => void;
  toggleSidebar: () => void;
  setSidebarCollapsed: (collapsed: boolean) => void;
  setSelectedFolder: (folderId: string) => void;
  setSelectedEmails: (emailIds: string[]) => void;
  toggleEmailSelection: (emailId: string) => void;
  clearEmailSelection: () => void;
  setViewMode: (mode: 'list' | 'grid') => void;
  setSortBy: (sortBy: 'date' | 'sender' | 'subject' | 'size') => void;
  setSortOrder: (order: 'asc' | 'desc') => void;
  setSearchQuery: (query: string) => void;
  setFilters: (filters: EmailFilters) => void;
  updateFilter: <K extends keyof EmailFilters>(key: K, value: EmailFilters[K]) => void;
  resetFilters: () => void;
}

const initialFilters: EmailFilters = {
  read: null,
  starred: null,
  hasAttachments: null,
  dateRange: {
    start: null,
    end: null,
  },
  labels: [],
};

export const useUIStore = create<UIStore>()(
  persist(
    (set) => ({
      // Initial state
      theme: 'system',
      sidebarCollapsed: false,
      selectedFolder: 'inbox',
      selectedEmails: [],
      viewMode: 'list',
      sortBy: 'date',
      sortOrder: 'desc',
      searchQuery: '',
      filters: initialFilters,

      // Actions
      setTheme: (theme) => {
        set({ theme });
        // Apply theme to document
        if (theme === 'dark' || (theme === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
          document.documentElement.classList.add('dark');
        } else {
          document.documentElement.classList.remove('dark');
        }
      },

      toggleSidebar: () => {
        set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed }));
      },

      setSidebarCollapsed: (collapsed) => {
        set({ sidebarCollapsed: collapsed });
      },

      setSelectedFolder: (folderId) => {
        set({ selectedFolder: folderId });
      },

      setSelectedEmails: (emailIds) => {
        set({ selectedEmails: emailIds });
      },

      toggleEmailSelection: (emailId) => {
        set((state) => {
          const isSelected = state.selectedEmails.includes(emailId);
          if (isSelected) {
            return {
              selectedEmails: state.selectedEmails.filter((id) => id !== emailId),
            };
          } else {
            return {
              selectedEmails: [...state.selectedEmails, emailId],
            };
          }
        });
      },

      clearEmailSelection: () => {
        set({ selectedEmails: [] });
      },

      setViewMode: (mode) => {
        set({ viewMode: mode });
      },

      setSortBy: (sortBy) => {
        set({ sortBy });
      },

      setSortOrder: (order) => {
        set({ sortOrder: order });
      },

      setSearchQuery: (query) => {
        set({ searchQuery: query });
      },

      setFilters: (filters) => {
        set({ filters });
      },

      updateFilter: (key, value) => {
        set((state) => ({
          filters: {
            ...state.filters,
            [key]: value,
          },
        }));
      },

      resetFilters: () => {
        set({ filters: initialFilters });
      },
    }),
    {
      name: 'ui-storage',
      partialize: (state) => ({
        theme: state.theme,
        sidebarCollapsed: state.sidebarCollapsed,
        viewMode: state.viewMode,
        sortBy: state.sortBy,
        sortOrder: state.sortOrder,
      }),
    }
  )
); 