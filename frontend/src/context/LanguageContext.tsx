import React, { createContext, useContext, useState, useEffect } from 'react';

type Language = 'en' | 'zh';

const translations = {
  en: {
    // Navigation
    'nav.market': 'Marketplace',
    'nav.companions': 'My Companions',
    'nav.chats': 'Chats',
    'nav.create': 'Create',
    'nav.profile': 'Profile',
    'nav.credits': 'Credits',
    'nav.topUp': 'Top Up',
    'nav.signOut': 'Sign Out',
    'nav.recentChats': 'Recent Chats',
    'nav.noChats': 'No chats yet',
    'nav.deleteChat': 'Delete Chat',
    'nav.confirmDelete': 'Are you sure you want to delete this conversation?',

    // Market
    'market.title': 'Community Market',
    'market.subtitle': 'Discover and fork characters created by the community.',
    'market.fork': 'Add to My List',
    'market.searchPlaceholder': 'Search market...',

    // Landing
    'landing.hero.title': 'Craft Your Digital',
    'landing.hero.highlight': 'Soulmate',
    'landing.hero.subtitle': 'Enter a universe of limitless AI companions. Create, share, and chat with characters that understand you.',
    'landing.cta.primary': 'Get Started',
    'landing.cta.secondary': 'View Market',
    'landing.feature.1.title': 'Unique Personalities',
    'landing.feature.1.desc': 'Every character has their own memory, emotions, and backstory.',
    'landing.feature.2.title': 'Global Marketplace',
    'landing.feature.2.desc': 'Discover and fork characters created by a vibrant community.',
    'landing.feature.3.title': 'Immersive Chat',
    'landing.feature.3.desc': 'Experience real-time, emotionally responsive conversations.',
    
    // Auth
    'auth.welcome': 'Welcome Back',
    'auth.subtitle': 'Sign in to continue your journey with your AI companions.',
    'auth.email': 'Email',
    'auth.password': 'Password',
    'auth.signIn': 'Sign In',
    'auth.mockCreds': 'Mock Credentials: user@example.com / password',
    'auth.failed': 'Login failed. Please try again.',
    'auth.emailOrPhone': 'Email or Phone Number',
    'auth.verifyTitle': 'Security Verification',
    'auth.verifyDesc': 'To continue, please enter the verification code sent to your account.',
    'auth.sendCode': 'Send Code',
    'auth.verify': 'Verify',
    'auth.code': 'Verification Code',
    'auth.codeSent': 'Code sent!',
    'auth.codeInvalid': 'Invalid code. Please try again.',

    // Dashboard
    'dash.title': 'Explore Companions',
    'dash.subtitle': 'Choose an AI character to start a conversation with.',
    'dash.createNew': 'Create New',
    'dash.startChat': 'Start Chat',
    'dash.searchPlaceholder': 'Search your companions...',

    // Chat
    'chat.online': 'Online',
    'chat.typePlaceholder': 'Type your message...',
    'chat.aiDisclaimer': 'AI generates responses. Credits deducted per message.',
    'chat.noActive': 'No active chats',
    'chat.startPrompt': 'Start a conversation with an AI companion from the dashboard.',
    'chat.exploreBtn': 'Explore Companions',

    // Create
    'create.title': 'Create New Character',
    'create.subtitle': 'Design your perfect AI companion. Define their personality and backstory.',
    'create.uploadAvatar': 'Click to upload avatar',
    'create.name': 'Name',
    'create.description': 'Description',
    'create.personality': 'Personality',
    'create.tags': 'Tags (comma separated)',
    'create.submit': 'Create Character',
    'create.save': 'Save Changes',
    'create.editTitle': 'Edit Character',
    'create.editSubtitle': 'Update your companion\'s details.',
    'create.placeholder.name': 'e.g. Dr. Watson',
    'create.placeholder.desc': 'Short bio...',
    'create.placeholder.pers': 'e.g. Helpful, Sarcastic, Professional...',
    'create.placeholder.tags': 'e.g. Helper, Coding, Medical',

    // Profile
    'profile.title': 'Your Profile',
    'profile.premium': 'Premium Member',
    'profile.balance': 'Credit Balance',
    'profile.autoRefill': 'Auto-refill OFF',
    'profile.remaining': 'messages remaining',
    'profile.buy': 'Buy Credits',
    'profile.history': 'View History',
    'profile.recentActivity': 'Recent Activity',
    'profile.sessionWith': 'Chat session with Character',
    'profile.ago': 'hours ago',
    'profile.tab.overview': 'Overview',
    'profile.tab.memories': 'Memories',
    'profile.tab.settings': 'Settings',
    'profile.memories.title': 'Memory Persona',
    'profile.memories.desc': 'Facts the AI has learned about you to personalize conversations.',
    'profile.memories.empty': 'No memories recorded yet.',
    'profile.settings.password': 'Change Password',
    'profile.settings.oldPass': 'Current Password',
    'profile.settings.newPass': 'New Password',
    'profile.settings.update': 'Update Password',
    'profile.settings.danger': 'Danger Zone',
    'profile.settings.deleteAccount': 'Delete Account',
    'profile.settings.deleteDesc': 'Permanently delete your account and all data. This action cannot be undone.',
    'profile.settings.confirmDelete': 'Are you sure? This action is irreversible.',

    // Tags
    'tag.assistant': 'Assistant',
    'tag.therapy': 'Therapy',
    'tag.coding': 'Coding',
    'tag.sci-fi': 'Sci-Fi',
    'tag.fantasy': 'Fantasy',
    'tag.history': 'History',
    'tag.anime': 'Anime',
    'tag.gaming': 'Gaming',
    'tag.education': 'Education',
    'tag.language-learning': 'Language Learning',
    'tag.creative-writing': 'Creative Writing',
    'tag.roleplay': 'Roleplay',
    'tag.philosophy': 'Philosophy',
    'tag.business': 'Business',
    'tag.wellness': 'Wellness',
    'tag.tech': 'Tech',
    'tag.wisdom': 'Wisdom',
    'tag.support': 'Support',
  },
  zh: {
    // Navigation
    'nav.market': '发现市场',
    'nav.companions': '我的伙伴',
    'nav.chats': '会话列表',
    'nav.create': '创建角色',
    'nav.profile': '个人中心',
    'nav.credits': '积分余额',
    'nav.topUp': '充值',
    'nav.signOut': '退出登录',
    'nav.recentChats': '最近会话',
    'nav.noChats': '暂无会话',
    'nav.deleteChat': '删除会话',
    'nav.confirmDelete': '您确定要删除这个会话吗？',

    // Market
    'market.title': '社区市场',
    'market.subtitle': '发现并复制来自社区的AI角色。',
    'market.fork': '添加到我的列表',
    'market.searchPlaceholder': '搜索市场...',

    // Landing
    'landing.hero.title': '打造你的数字',
    'landing.hero.highlight': '灵魂伴侣',
    'landing.hero.subtitle': '进入无限的AI伙伴宇宙。创造、分享并与懂你的角色畅聊。',
    'landing.cta.primary': '立即开始',
    'landing.cta.secondary': '浏览市场',
    'landing.feature.1.title': '独一无二的个性',
    'landing.feature.1.desc': '每个角色都拥有独立的记忆、情感和背景故事。',
    'landing.feature.2.title': '全球社区市场',
    'landing.feature.2.desc': '发现并复制来自活跃社区的精彩角色。',
    'landing.feature.3.title': '沉浸式对话',
    'landing.feature.3.desc': '体验实时、具有情感反馈的深度交流。',

    // Auth
    'auth.welcome': '欢迎回来',
    'auth.subtitle': '登录以继续您与AI伙伴的旅程。',
    'auth.email': '邮箱',
    'auth.password': '密码',
    'auth.signIn': '登录',
    'auth.mockCreds': '演示账号: user@example.com / password',
    'auth.failed': '登录失败，请重试。',
    'auth.emailOrPhone': '邮箱或手机号',
    'auth.verifyTitle': '安全验证',
    'auth.verifyDesc': '为了继续操作，请输入发送到您账号的验证码。',
    'auth.sendCode': '获取验证码',
    'auth.verify': '验证',
    'auth.code': '验证码',
    'auth.codeSent': '验证码已发送！',
    'auth.codeInvalid': '验证码无效，请重试。',

    // Dashboard
    'dash.title': '发现伙伴',
    'dash.subtitle': '选择一个AI角色开始对话。',
    'dash.createNew': '创建新角色',
    'dash.startChat': '开始对话',
    'dash.searchPlaceholder': '搜索我的伙伴...',

    // Chat
    'chat.online': '在线',
    'chat.typePlaceholder': '输入消息...', 
    'chat.aiDisclaimer': 'AI生成内容。每条消息扣除积分。',
    'chat.noActive': '暂无活跃会话',
    'chat.startPrompt': '从仪表盘开始与AI伙伴的对话。',
    'chat.exploreBtn': '去发现伙伴',

    // Create
    'create.title': '创建新角色',
    'create.subtitle': '设计您完美的AI伙伴。定义他们的性格和背景故事。',
    'create.uploadAvatar': '点击上传头像',
    'create.name': '名称',
    'create.description': '简介',
    'create.personality': '性格设定',
    'create.tags': '标签 (逗号分隔)',
    'create.submit': '创建角色',
    'create.save': '保存修改',
    'create.editTitle': '编辑角色',
    'create.editSubtitle': '更新您的伙伴详情。',
    'create.placeholder.name': '例如：华生医生',
    'create.placeholder.desc': '简短的个人简介...', 
    'create.placeholder.pers': '例如：乐于助人、幽默、专业...',
    'create.placeholder.tags': '例如：助手, 编程, 医疗',

    // Profile
    'profile.title': '个人资料',
    'profile.premium': '高级会员',
    'profile.balance': '积分余额',
    'profile.autoRefill': '自动充值 已关闭',
    'profile.remaining': '条剩余消息',
    'profile.buy': '购买积分',
    'profile.history': '查看记录',
    'profile.recentActivity': '最近活动',
    'profile.sessionWith': '与角色的会话',
    'profile.ago': '小时前',
    'profile.tab.overview': '概览',
    'profile.tab.memories': '记忆管理',
    'profile.tab.settings': '账号设置',
    'profile.memories.title': '个人画像',
    'profile.memories.desc': 'AI 从对话中了解到的关于您的信息，用于提供个性化体验。',
    'profile.memories.empty': '暂无记忆记录。',
    'profile.settings.password': '修改密码',
    'profile.settings.oldPass': '当前密码',
    'profile.settings.newPass': '新密码',
    'profile.settings.update': '更新密码',
    'profile.settings.danger': '危险区域',
    'profile.settings.deleteAccount': '删除账号',
    'profile.settings.deleteDesc': '永久删除您的账号及所有数据。此操作无法撤销。',
    'profile.settings.confirmDelete': '您确定吗？此操作不可逆。',

    // Tags
    'tag.assistant': '助手',
    'tag.therapy': '心理咨询',
    'tag.coding': '编程',
    'tag.sci-fi': '科幻',
    'tag.fantasy': '奇幻',
    'tag.history': '历史',
    'tag.anime': '动漫',
    'tag.gaming': '游戏',
    'tag.education': '教育',
    'tag.language-learning': '语言学习',
    'tag.creative-writing': '创意写作',
    'tag.roleplay': '角色扮演',
    'tag.philosophy': '哲学',
    'tag.business': '商业',
    'tag.wellness': '健康',
    'tag.tech': '科技',
    'tag.wisdom': '智慧',
    'tag.support': '情感支持',
  }
};

interface LanguageContextType {
  language: Language;
  setLanguage: (lang: Language) => void;
  t: (key: keyof typeof translations.en) => string;
}

const LanguageContext = createContext<LanguageContextType | undefined>(undefined);

export const LanguageProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [language, setLanguage] = useState<Language>(() => {
    return (localStorage.getItem('language') as Language) || 'zh';
  });

  useEffect(() => {
    localStorage.setItem('language', language);
  }, [language]);

  const t = (key: keyof typeof translations.en) => {
    return translations[language][key] || key;
  };

  return (
    <LanguageContext.Provider value={{ language, setLanguage, t }}>
      {children}
    </LanguageContext.Provider>
  );
};

export const useLanguage = () => {
  const context = useContext(LanguageContext);
  if (context === undefined) {
    throw new Error('useLanguage must be used within a LanguageProvider');
  }
  return context;
};