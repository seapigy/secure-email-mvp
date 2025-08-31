import { create } from 'zustand';
import { log } from '@/lib/logger';

interface MetricsData {
  total_links: number;
  active_links: number;
  total_views: number;
  failed_attempts: number;
  dlp_scans: number;
  security_violations: number;
  storage_used: number;
  storage_limit: number;
  [key: string]: number; // Allow additional metrics
}

interface HealthData {
  status: 'healthy' | 'degraded' | 'unhealthy';
  services: {
    database: { status: 'up' | 'down'; response_time?: number };
    storage: { status: 'up' | 'down'; response_time?: number };
    dlp: { status: 'up' | 'down'; response_time?: number };
    auth: { status: 'up' | 'down'; response_time?: number };
  };
  last_check: string;
  [key: string]: unknown; // Allow additional health data
}

interface StreamEvent {
  type: string;
  data: MetricsData | HealthData | unknown;
  timestamp: string;
}

interface MonitoringState {
  // Connection state
  isConnected: boolean;
  connectionError: string | null;
  
  // Real-time metrics
  currentMetrics: MetricsData | null;
  systemHealth: HealthData | null;
  
  // Event stream
  events: StreamEvent[];
  subscribers: Map<string, (event: StreamEvent) => void>;
  
  // SSE connection
  eventSource: EventSource | null;
  
  // Actions
  connect: () => void;
  disconnect: () => void;
  subscribeToStream: (clientId: string, callback: (event: StreamEvent) => void) => void;
  unsubscribeFromStream: (clientId: string) => void;
  addEvent: (event: StreamEvent) => void;
  clearEvents: () => void;
  updateMetrics: (metrics: MetricsData) => void;
  updateSystemHealth: (health: HealthData) => void;
}

export const useMonitoringStore = create<MonitoringState>((set, get) => ({
  // Initial state
  isConnected: false,
  connectionError: null,
  currentMetrics: null,
  systemHealth: null,
  events: [],
  subscribers: new Map(),
  eventSource: null,

  // Connect to SSE stream
  connect: () => {
    const { disconnect } = get();
    
    // Disconnect existing connection
    disconnect();
    
    try {
      const eventSource = new EventSource('/api/metrics/stream');
      
      eventSource.onopen = () => {
        set({ 
          isConnected: true, 
          connectionError: null,
          eventSource 
        });
        log.info('Monitoring SSE connection established', null, 'monitoringStore');
      };
      
      eventSource.onmessage = (event) => {
        try {
          const streamEvent: StreamEvent = JSON.parse(event.data);
          const { addEvent, subscribers } = get();
          
          // Add to events list
          addEvent(streamEvent);
          
          // Notify all subscribers
          subscribers.forEach((callback) => {
            try {
              callback(streamEvent);
            } catch (error) {
              log.error('Error in monitoring subscriber callback:', error, 'monitoringStore');
            }
          });
          
          // Update specific state based on event type
          if (streamEvent.type === 'metrics_update' || streamEvent.type === 'initial_metrics') {
            get().updateMetrics(streamEvent.data as MetricsData);
          } else if (streamEvent.type === 'health_update') {
            get().updateSystemHealth(streamEvent.data as HealthData);
          }
          
        } catch (error) {
          log.error('Error parsing SSE event:', error, 'monitoringStore');
        }
      };
      
      eventSource.onerror = (error) => {
        log.error('SSE connection error:', error, 'monitoringStore');
        set({ 
          isConnected: false, 
          connectionError: 'Connection failed',
          eventSource: null 
        });
        
        // Attempt to reconnect after 5 seconds
        setTimeout(() => {
          const { isConnected } = get();
          if (!isConnected) {
            get().connect();
          }
        }, 5000);
      };
      
    } catch (error) {
              log.error('Failed to create SSE connection:', error, 'monitoringStore');
      set({ 
        isConnected: false, 
        connectionError: 'Failed to establish connection',
        eventSource: null 
      });
    }
  },

  // Disconnect from SSE stream
  disconnect: () => {
    const { eventSource } = get();
    if (eventSource) {
      eventSource.close();
    }
    set({ 
      isConnected: false, 
      eventSource: null,
      connectionError: null 
    });
  },

  // Subscribe to stream events
  subscribeToStream: (clientId: string, callback: (event: StreamEvent) => void) => {
    const { subscribers, connect } = get();
    const newSubscribers = new Map(subscribers);
    newSubscribers.set(clientId, callback);
    
    set({ subscribers: newSubscribers });
    
    // Auto-connect if not already connected
    if (!get().isConnected) {
      connect();
    }
  },

  // Unsubscribe from stream events
  unsubscribeFromStream: (clientId: string) => {
    const { subscribers } = get();
    const newSubscribers = new Map(subscribers);
    newSubscribers.delete(clientId);
    set({ subscribers: newSubscribers });
    
    // Disconnect if no more subscribers
    if (newSubscribers.size === 0) {
      get().disconnect();
    }
  },

  // Add event to events list
  addEvent: (event: StreamEvent) => {
    const { events } = get();
    const newEvents = [...events, event];
    
    // Keep only last 100 events
    if (newEvents.length > 100) {
      newEvents.splice(0, newEvents.length - 100);
    }
    
    set({ events: newEvents });
  },

  // Clear events
  clearEvents: () => {
    set({ events: [] });
  },

  // Update current metrics
  updateMetrics: (metrics: MetricsData) => {
    set({ currentMetrics: metrics });
  },

  // Update system health
  updateSystemHealth: (health: HealthData) => {
    set({ systemHealth: health });
  },
}));

// Auto-connect when store is first used
let isInitialized = false;
export const initializeMonitoring = () => {
  if (!isInitialized) {
    isInitialized = true;
    // Don't auto-connect - let components decide when to connect
  }
};

// Export selectors for better performance
export const useMonitoringConnection = () => useMonitoringStore((state) => ({
  isConnected: state.isConnected,
  connectionError: state.connectionError,
  connect: state.connect,
  disconnect: state.disconnect,
}));

export const useMonitoringMetrics = () => useMonitoringStore((state) => ({
  currentMetrics: state.currentMetrics,
  systemHealth: state.systemHealth,
  updateMetrics: state.updateMetrics,
  updateSystemHealth: state.updateSystemHealth,
}));

export const useMonitoringEvents = () => useMonitoringStore((state) => ({
  events: state.events,
  addEvent: state.addEvent,
  clearEvents: state.clearEvents,
}));
