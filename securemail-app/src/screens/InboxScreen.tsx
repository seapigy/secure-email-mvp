// Inbox Screen for SecureMail
import React, { useState, useEffect } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TouchableOpacity,
  FlatList,
  RefreshControl,
} from 'react-native';
import { useAuth } from '../contexts/AuthContext';
import { apiService } from '../services/api';
import { InboxFolder, EmailMessage } from '../types';

export default function InboxScreen() {
  const { state } = useAuth();
  const [folders, setFolders] = useState<InboxFolder[]>([]);
  const [messages, setMessages] = useState<EmailMessage[]>([]);
  const [selectedFolder, setSelectedFolder] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);

  useEffect(() => {
    if (state.isAuthenticated && state.token) {
      loadFolders();
    }
  }, [state.isAuthenticated, state.token]);

  useEffect(() => {
    if (selectedFolder && state.token) {
      loadMessages(selectedFolder);
    }
  }, [selectedFolder, state.token]);

  const loadFolders = async () => {
    if (!state.token) return;
    
    try {
      const response = await apiService.getInboxFolders(state.token);
      setFolders(response.folders);
      
      // Select inbox folder by default
      const inboxFolder = response.folders.find(f => f.folderType === 'inbox');
      if (inboxFolder) {
        setSelectedFolder(inboxFolder.id);
      }
    } catch (error) {
      console.error('Error loading folders:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const loadMessages = async (folderId: string) => {
    if (!state.token) return;
    
    try {
      const response = await apiService.getInboxMessages(state.token, folderId);
      setMessages(response.messages);
    } catch (error) {
      console.error('Error loading messages:', error);
    }
  };

  const handleRefresh = async () => {
    setIsRefreshing(true);
    try {
      await loadFolders();
      if (selectedFolder) {
        await loadMessages(selectedFolder);
      }
    } finally {
      setIsRefreshing(false);
    }
  };

  const renderFolder = ({ item }: { item: InboxFolder }) => (
    <TouchableOpacity
      style={[
        styles.folderItem,
        selectedFolder === item.id && styles.folderItemSelected,
      ]}
      onPress={() => setSelectedFolder(item.id)}
    >
      <Text style={[
        styles.folderName,
        selectedFolder === item.id && styles.folderNameSelected,
      ]}>
        {item.name}
      </Text>
    </TouchableOpacity>
  );

  const renderMessage = ({ item }: { item: EmailMessage }) => (
    <TouchableOpacity style={styles.messageItem}>
      <View style={styles.messageHeader}>
        <Text style={styles.messageFrom} numberOfLines={1}>
          {item.from}
        </Text>
        <Text style={styles.messageDate}>
          {new Date(item.receivedAt).toLocaleDateString()}
        </Text>
      </View>
      <Text style={styles.messageSubject} numberOfLines={1}>
        {item.subject}
      </Text>
      <View style={styles.messageFooter}>
        <Text style={styles.messageSize}>
          {Math.round(item.sizeBytes / 1024)} KB
        </Text>
        {item.isImportant && <Text style={styles.importantFlag}>⭐</Text>}
        {item.isStarred && <Text style={styles.starredFlag}>⭐</Text>}
      </View>
    </TouchableOpacity>
  );

  if (isLoading) {
    return (
      <View style={styles.loadingContainer}>
        <Text style={styles.loadingText}>Loading inbox...</Text>
      </View>
    );
  }

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.title}>Inbox</Text>
      </View>

      <View style={styles.content}>
        <View style={styles.foldersContainer}>
          <FlatList
            data={folders}
            renderItem={renderFolder}
            keyExtractor={(item) => item.id}
            horizontal
            showsHorizontalScrollIndicator={false}
            contentContainerStyle={styles.foldersList}
          />
        </View>

        <View style={styles.messagesContainer}>
          {selectedFolder ? (
            <FlatList
              data={messages}
              renderItem={renderMessage}
              keyExtractor={(item) => item.id}
              refreshControl={
                <RefreshControl
                  refreshing={isRefreshing}
                  onRefresh={handleRefresh}
                />
              }
              ListEmptyComponent={
                <View style={styles.emptyContainer}>
                  <Text style={styles.emptyText}>No messages in this folder</Text>
                </View>
              }
            />
          ) : (
            <View style={styles.emptyContainer}>
              <Text style={styles.emptyText}>Select a folder to view messages</Text>
            </View>
          )}
        </View>
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#F5F5F5',
  },
  loadingContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#F5F5F5',
  },
  loadingText: {
    fontSize: 16,
    color: '#666666',
  },
  header: {
    backgroundColor: '#FFFFFF',
    padding: 20,
    borderBottomWidth: 1,
    borderBottomColor: '#E0E0E0',
  },
  title: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#333333',
  },
  content: {
    flex: 1,
  },
  foldersContainer: {
    backgroundColor: '#FFFFFF',
    borderBottomWidth: 1,
    borderBottomColor: '#E0E0E0',
  },
  foldersList: {
    paddingHorizontal: 20,
    paddingVertical: 12,
  },
  folderItem: {
    paddingHorizontal: 16,
    paddingVertical: 8,
    marginRight: 12,
    borderRadius: 20,
    backgroundColor: '#F0F0F0',
  },
  folderItemSelected: {
    backgroundColor: '#007AFF',
  },
  folderName: {
    fontSize: 14,
    color: '#333333',
    fontWeight: '500',
  },
  folderNameSelected: {
    color: '#FFFFFF',
  },
  messagesContainer: {
    flex: 1,
  },
  messageItem: {
    backgroundColor: '#FFFFFF',
    padding: 16,
    borderBottomWidth: 1,
    borderBottomColor: '#F0F0F0',
  },
  messageHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: 8,
  },
  messageFrom: {
    fontSize: 16,
    fontWeight: '600',
    color: '#333333',
    flex: 1,
  },
  messageDate: {
    fontSize: 12,
    color: '#666666',
  },
  messageSubject: {
    fontSize: 14,
    color: '#333333',
    marginBottom: 8,
  },
  messageFooter: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  messageSize: {
    fontSize: 12,
    color: '#666666',
  },
  importantFlag: {
    fontSize: 16,
  },
  starredFlag: {
    fontSize: 16,
  },
  emptyContainer: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    padding: 40,
  },
  emptyText: {
    fontSize: 16,
    color: '#666666',
    textAlign: 'center',
  },
});
