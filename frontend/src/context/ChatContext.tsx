import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';
import type { ChatSession } from '../types';
import { mockApi } from '../services/mockApi';
import { useAuth } from './AuthContext';

interface ChatContextType {
  sessions: ChatSession[];
  isLoading: boolean;
  createSession: (characterId: string) => Promise<string>;
  deleteSession: (sessionId: string) => Promise<void>;
  refreshSessions: () => Promise<void>;
}

const ChatContext = createContext<ChatContextType | undefined>(undefined);

export const ChatProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { user } = useAuth();
  const [sessions, setSessions] = useState<ChatSession[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  const refreshSessions = useCallback(async () => {
    if (!user) {
      setSessions([]);
      return;
    }
    // Don't set global loading here to avoid flickering sidebar
    try {
      const data = await mockApi.chat.getSessions();
      setSessions(data);
    } catch (error) {
      console.error("Failed to fetch sessions", error);
    }
  }, [user]);

  // Initial load
  useEffect(() => {
    refreshSessions();
  }, [refreshSessions]);

  const createSession = async (characterId: string): Promise<string> => {
    setIsLoading(true);
    try {
      const newSession = await mockApi.chat.createSession(characterId);
      // Optimistically update list immediately
      setSessions(prev => [newSession, ...prev]); 
      return newSession.id;
    } finally {
      setIsLoading(false);
    }
  };

  const deleteSession = async (sessionId: string): Promise<void> => {
    // Optimistic update
    setSessions(prev => prev.filter(s => s.id !== sessionId));
    try {
      await mockApi.chat.deleteSession(sessionId);
      // No need to refresh, we already removed it
    } catch (error) {
      console.error("Failed to delete session", error);
      // Revert if failed (optional, but good practice)
      refreshSessions();
    }
  };

  return (
    <ChatContext.Provider value={{ sessions, isLoading, createSession, deleteSession, refreshSessions }}>
      {children}
    </ChatContext.Provider>
  );
};

export const useChat = () => {
  const context = useContext(ChatContext);
  if (context === undefined) {
    throw new Error('useChat must be used within a ChatProvider');
  }
  return context;
};
