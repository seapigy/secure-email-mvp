import { create } from 'zustand';

interface StreamEvent {
  type: string;
  data: any;
  timestamp: string;
}

interface MonitoringState {
  // Connection state
  isConnected: boolean;
  connectionError: string | null;
  
  // Real-time metrics
  currentMetrics: any | null;
  systemHealth: any | null;
  
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
  updateMetrics: (metrics: any) => void;
  updateSystemHealth: (health: any) => void;
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
        console.log('Monitoring SSE connection established');
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
              console.error('Error in monitoring subscriber callback:', error);
            }
          });
          
          // Update specific state based on event type
          if (streamEvent.type === 'metrics_update' || streamEvent.type === 'initial_metrics') {
            get().updateMetrics(streamEvent.data);
          } else if (streamEvent.type === 'health_update') {
            get().updateSystemHealth(streamEvent.data);
          }
          
        } catch (error) {
          console.error('Error parsing SSE event:', error);
        }
      };
      
      eventSource.onerror = (error) => {
        console.error('SSE connection error:', error);
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
      console.error('Failed to create SSE connection:', error);
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
  updateMetrics: (metrics: any) => {
    set({ currentMetrics: metrics });
  },

  // Update system health
  updateSystemHealth: (health: any) => {
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
