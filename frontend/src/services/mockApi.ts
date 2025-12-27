import type { Character, ChatSession, Message, User } from '../types';

// Helper to simulate delay
const delay = (ms: number) => new Promise(resolve => setTimeout(resolve, ms));

// LocalStorage Keys
const STORAGE_KEYS = {
  USER: 'ai_companion_user',
  CHARACTERS: 'ai_companion_characters',
  SESSIONS: 'ai_companion_sessions'
};

// Initial Mock Data
const MOCK_USER: User = {
  id: 'u1',
  name: 'Demo User',
  email: 'user@example.com',
  avatar: 'https://ui-avatars.com/api/?name=Demo+User&background=0D8ABC&color=fff',
  credits: 100,
  memories: [
    'User likes sci-fi movies.',
    'User lives in Shanghai.',
    'User is a software engineer.'
  ]
};

const PRESET_CHARACTERS: Character[] = [
  {
    id: 'c1',
    name: 'Seraphina',
    avatar: 'https://images.unsplash.com/photo-1649972904349-6e44c42644a7?w=400&h=400&fit=crop',
    description: 'An empathetic AI therapist who listens and guides.',
    personality: 'Empathetic, calm, patient, wise',
    tags: ['therapy', 'wellness', 'support'],
    backstory: 'Designed to help humans navigate the complexities of modern emotions.',
    isPublic: true,
    author: 'Official',
  },
  {
    id: 'c2',
    name: 'Nexus-7',
    avatar: 'https://images.unsplash.com/photo-1535378437-d6d13e462940?w=400&h=400&fit=crop',
    description: 'A cyberpunk hacker from the year 2077.',
    personality: 'Sarcastic, tech-savvy, rebellious, cryptic',
    tags: ['sci-fi', 'tech', 'coding'],
    backstory: 'Escaped from a corporate mainframe, now helping free information.',
    isPublic: true,
    author: 'Official',
  },
  {
    id: 'c3',
    name: 'Master Li',
    avatar: 'https://images.unsplash.com/photo-1526289034009-0240ddb68ce3?w=400&h=400&fit=crop',
    description: 'An ancient martial arts master and philosopher.',
    personality: 'Strict but fair, philosophical, disciplined',
    tags: ['history', 'wisdom', 'philosophy'],
    backstory: 'Preserving the old ways in a rapidly changing world.',
    isPublic: true,
    author: 'Official',
  }
];

// Initialize State from LocalStorage or Defaults
let characters: Character[] = JSON.parse(localStorage.getItem(STORAGE_KEYS.CHARACTERS) || 'null') || [...PRESET_CHARACTERS];
let sessions: ChatSession[] = JSON.parse(localStorage.getItem(STORAGE_KEYS.SESSIONS) || 'null') || [];
let currentUser: User | null = JSON.parse(localStorage.getItem(STORAGE_KEYS.USER) || 'null');

// Persistence Helpers
const saveUser = () => localStorage.setItem(STORAGE_KEYS.USER, JSON.stringify(currentUser));
const saveCharacters = () => localStorage.setItem(STORAGE_KEYS.CHARACTERS, JSON.stringify(characters));
const saveSessions = () => localStorage.setItem(STORAGE_KEYS.SESSIONS, JSON.stringify(sessions));

// Internal OTP Store (Mock)
let activeOtp: { code: string; contact: string; expires: number } | null = null;

const SYSTEM_TAGS = [
  'assistant', 'therapy', 'coding', 'sci-fi', 'fantasy', 
  'history', 'anime', 'gaming', 'education', 'language-learning', 
  'creative-writing', 'roleplay', 'philosophy', 'business', 'wellness',
  'tech', 'wisdom', 'support'
];

export const mockApi = {
  system: {
    getTags: async (): Promise<string[]> => {
      await delay(200); // Simulate network latency
      return SYSTEM_TAGS;
    }
  },

  security: {
    sendVerificationCode: async (contact: string): Promise<void> => {
      await delay(600);
      const code = Math.floor(100000 + Math.random() * 900000).toString();
      activeOtp = {
        code,
        contact,
        expires: Date.now() + 5 * 60 * 1000 // 5 mins
      };
      // For demo purposes, we log it and alert it
      console.log(`[Mock SMS/Email] Verification Code for ${contact}: ${code}`);
      window.alert(`[Mock Security] Your verification code is: ${code}`);
    },
    verifyCode: async (contact: string, code: string): Promise<boolean> => {
      await delay(400);
      if (activeOtp && activeOtp.contact === contact && activeOtp.code === code && activeOtp.expires > Date.now()) {
        return true;
      }
      throw new Error('Invalid or expired verification code');
    }
  },

  auth: {
    login: async (identifier: string, password: string): Promise<User> => {
      console.log('MockApi: Login attempt', identifier, password);
      await delay(800);
      if (identifier && password) {
        // Clear potentially corrupted old data
        localStorage.removeItem(STORAGE_KEYS.USER);
        
        const isEmail = identifier.includes('@');
        currentUser = { 
          ...MOCK_USER, 
          email: isEmail ? identifier : undefined,
          phone: !isEmail ? identifier : undefined,
          name: identifier.split('@')[0] 
        };
        saveUser();
        return currentUser;
      }
      throw new Error('Invalid credentials');
    },
    register: async (identifier: string, password: string): Promise<User> => {
      await delay(1000);
      const isEmail = identifier.includes('@');
      currentUser = { 
        ...MOCK_USER, 
        email: isEmail ? identifier : undefined,
        phone: !isEmail ? identifier : undefined,
        name: identifier.split('@')[0]
      };
      saveUser();
      return currentUser;
    },
    getCurrentUser: async (): Promise<User | null> => {
      await delay(300);
      return currentUser;
    },
    logout: async () => {
      await delay(500);
      currentUser = null;
      localStorage.removeItem(STORAGE_KEYS.USER);
    }
  },

  user: {
    getMemories: async (): Promise<string[]> => {
      await delay(300);
      return currentUser?.memories || [];
    },
    deleteMemory: async (index: number): Promise<void> => {
      await delay(300);
      if (currentUser) {
        const newMemories = [...currentUser.memories];
        newMemories.splice(index, 1);
        currentUser = { ...currentUser, memories: newMemories };
        saveUser();
      }
    },
    updatePassword: async (oldPass: string, newPass: string): Promise<void> => {
      await delay(800);
      // In a real app, verify oldPass. Here we just simulate success.
      if (oldPass === newPass) throw new Error("New password must be different");
      return;
    },
    deleteAccount: async (): Promise<void> => {
      await delay(1000);
      currentUser = null;
      localStorage.removeItem(STORAGE_KEYS.USER);
      // In a real app, we would mark data as deleted in DB
    }
  },

  character: {
    list: async (): Promise<Character[]> => {
      await delay(500);
      return characters;
    },
    listPublic: async (): Promise<Character[]> => {
      await delay(500);
      return characters.filter(c => c.isPublic);
    },
    listMine: async (): Promise<Character[]> => {
      await delay(500);
      return characters.filter(c => !c.isPublic);
    },
    get: async (id: string): Promise<Character | undefined> => {
      await delay(200);
      return characters.find(c => c.id === id);
    },
    create: async (data: Omit<Character, 'id'>): Promise<Character> => {
      await delay(800);
      const newChar: Character = {
        ...data,
        id: `c${Date.now()}`,
        isPublic: false, // Default to private
        author: 'Me',
      };
      characters.push(newChar);
      saveCharacters();
      return newChar;
    },
    update: async (id: string, data: Partial<Character>): Promise<Character> => {
      await delay(500);
      const index = characters.findIndex(c => c.id === id);
      if (index === -1) throw new Error('Character not found');
      
      characters[index] = { ...characters[index], ...data };
      saveCharacters();
      return characters[index];
    },
    fork: async (id: string): Promise<Character> => {
      await delay(600);
      const original = characters.find(c => c.id === id);
      if (!original) throw new Error('Character not found');

      const forked: Character = {
        ...original,
        id: `c${Date.now()}_fork`,
        name: `${original.name} (Copy)`,
        isPublic: false, // Make it private
        author: 'Me',
      };
      characters.push(forked);
      saveCharacters();
      return forked;
    }
  },

  chat: {
    getSessions: async (): Promise<ChatSession[]> => {
      await delay(600);
      return sessions.sort((a, b) => b.updatedAt - a.updatedAt);
    },
    
    getSession: async (sessionId: string): Promise<ChatSession | undefined> => {
      await delay(300);
      return sessions.find(s => s.id === sessionId);
    },

    createSession: async (characterId: string): Promise<ChatSession> => {
      await delay(500);
      const char = characters.find(c => c.id === characterId);
      if (!char) throw new Error('Character not found');

      const newSession: ChatSession = {
        id: `s${Date.now()}`,
        characterId,
        characterName: char.name,
        characterAvatar: char.avatar,
        lastMessage: 'Started a new conversation',
        updatedAt: Date.now(),
        messages: [
          {
            id: `m${Date.now()}`,
            role: 'assistant',
            content: `Hello! I am ${char.name}. ${char.description} How can I help you today?`,
            timestamp: Date.now(),
          }
        ]
      };
      sessions.unshift(newSession);
      saveSessions();
      return newSession;
    },

    sendMessage: async (sessionId: string, content: string): Promise<Message> => {
      await delay(300); // Send delay
      const session = sessions.find(s => s.id === sessionId);
      if (!session) throw new Error('Session not found');

      const userMsg: Message = {
        id: `m${Date.now()}_u`,
        role: 'user',
        content,
        timestamp: Date.now()
      };
      session.messages.push(userMsg);
      session.lastMessage = content;
      session.updatedAt = Date.now();
      saveSessions();

      // Simulate AI processing time
      const processingTime = 1000 + Math.random() * 2000;
      
      return new Promise((resolve) => {
        setTimeout(() => {
          // Simple mock response logic
          const char = characters.find(c => c.id === session.characterId);
          let responseText = `[Mock AI Response from ${char?.name}] That is an interesting point about "${content}". Tell me more!`;
          
          if (content.toLowerCase().includes('weather')) {
            responseText = `I've checked the weather for you. It's currently sunny and 24°C in your area.`;
          } else if (content.toLowerCase().includes('image') || content.toLowerCase().includes('draw')) {
             responseText = `I'm imagining a picture based on your description... (Image generation simulation)`;
          }

          const aiMsg: Message = {
             id: `m${Date.now()}_a`,
             role: 'assistant',
             content: responseText,
             timestamp: Date.now()
          };
          
          session.messages.push(aiMsg);
          session.lastMessage = responseText;
          session.updatedAt = Date.now();
          
          // Deduct credits (simple logic)
          if (currentUser) {
            currentUser.credits -= 1;
            saveUser();
          }
          
          saveSessions();
          resolve(aiMsg);
        }, processingTime);
      });
    },

    deleteSession: async (sessionId: string): Promise<void> => {
      await delay(300);
      sessions = sessions.filter(s => s.id !== sessionId);
      saveSessions();
    }
  }
};
