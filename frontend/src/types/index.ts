export interface User {
  id: string;
  name: string;
  email?: string;
  phone?: string;
  avatar?: string;
  credits: number;
  memories: string[]; // List of facts AI remembers about the user
}

export interface Character {
  id: string;
  name: string;
  avatar: string;
  description: string;
  personality: string;
  backstory?: string;
  systemPrompt?: string; // Internal use for simulation
  tags: string[];
  isPublic?: boolean;
  author?: string;
}

export interface Message {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: number;
  toolCall?: {
    toolName: string;
    input: string;
    status: 'pending' | 'success' | 'error';
    result?: string;
  };
}

export interface ChatSession {
  id: string;
  characterId: string;
  characterName: string; // Cached for display
  characterAvatar: string; // Cached for display
  lastMessage: string;
  updatedAt: number;
  messages: Message[];
}

export interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}
