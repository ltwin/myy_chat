import React from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import { ThemeProvider } from './context/ThemeContext';
import { ChatProvider } from './context/ChatContext';
import { LanguageProvider } from './context/LanguageContext';
import { ProtectedRoute } from './components/layout/ProtectedRoute';
import { LoginPage } from './pages/LoginPage';
import { Dashboard } from './pages/Dashboard';
import { MarketplacePage } from './pages/MarketplacePage';
import { ChatPage } from './pages/ChatPage';
import { LandingPage } from './pages/LandingPage';
import { CreatePage } from './pages/CreatePage';
import { ProfilePage } from './pages/ProfilePage';

function App() {
  return (
    <BrowserRouter>
      <LanguageProvider>
        <ThemeProvider>
          <AuthProvider>
            <ChatProvider>
              <Routes>
                <Route path="/welcome" element={<LandingPage />} />
                <Route path="/login" element={<LoginPage />} />
                
                <Route element={<ProtectedRoute />}>
                  <Route path="/" element={<Dashboard />} />
                  <Route path="/market" element={<MarketplacePage />} />
                  <Route path="/chat" element={<ChatPage />} />
                  <Route path="/chat/:sessionId" element={<ChatPage />} />
                  <Route path="/create" element={<CreatePage />} />
                  <Route path="/edit/:charId" element={<CreatePage />} />
                  <Route path="/profile" element={<ProfilePage />} />
                </Route>

                <Route path="*" element={<Navigate to="/" replace />} />
              </Routes>
            </ChatProvider>
          </AuthProvider>
        </ThemeProvider>
      </LanguageProvider>
    </BrowserRouter>
  );
}

export default App;
