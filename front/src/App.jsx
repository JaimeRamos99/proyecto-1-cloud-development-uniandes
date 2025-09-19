import React, { useState, useEffect, useRef } from 'react';
import { Upload, Play, Trophy, User, Home, LogOut, ThumbsUp, Video, Star, TrendingUp, 
  CheckCircle, Clock, XCircle, Loader2, ChevronRight, Award, Users, MapPin, 
  Calendar, Eye, Filter, BarChart3, Shield, 
  ChevronLeft, RefreshCw} from 'lucide-react';
import apiService from './services/api';


const App = () => {
  const [currentView, setCurrentView] = useState('landing');
  const [user, setUser] = useState(null);
  const [selectedCity, setSelectedCity] = useState('todas');
  const [videos, setVideos] = useState([]);
  const [rankings, setRankings] = useState([]);
  const [myVideos, setMyVideos] = useState([]);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [processingStatus, setProcessingStatus] = useState(null);
  const [loading, setLoading] = useState(false);
  const [votedVideos, setVotedVideos] = useState(new Set());
  const [expandedVideo, setExpandedVideo] = useState(null);


  const cities = ['Todas', 'Bogotá', 'Medellín', 'Cali', 'Barranquilla', 'Cartagena', 'Bucaramanga', 'Pereira'];

  const [errorModal, setErrorModal] = useState({
    isOpen: false,
    title: '',
    message: '',
    technical: '',
    suggestions: []
  });
  // Check for existing auth on app load
  useEffect(() => {
    const checkAuth = async () => {
      if (apiService.isAuthenticated()) {
        try {
          const profile = await apiService.getProfile();
          setUser(profile);
          setCurrentView('dashboard');
        } catch (error) {
          console.error('Auth check failed:', error);
          apiService.logout();
        }
      }
    };
    checkAuth();
  }, []);

  // Load data based on current view
  useEffect(() => {
    const loadData = async () => {
      setLoading(true);
      try {
        if (currentView === 'videos' || currentView === 'dashboard') {
          const publicVideos = await apiService.getPublicVideos();
          setVideos(Array.isArray(publicVideos) ? publicVideos : []);
        }
        
        if (currentView === 'rankings' || currentView === 'dashboard') {
          const topRankings = await apiService.getTopRankings(50, selectedCity);
          // Backend returns { rankings: [...], pagination: {...} }
          const rankingsData = topRankings.rankings || topRankings;
          setRankings(Array.isArray(rankingsData) ? rankingsData : []);
        }

        if (currentView === 'dashboard' && user) {
          const userVideos = await apiService.getMyVideos();
          setMyVideos(Array.isArray(userVideos) ? userVideos : []);
        }
      } catch (error) {
        console.error('Failed to load data:', error);
        setVideos([]);
        setRankings([]);
        setMyVideos([]);
      } finally {
        setLoading(false);
      }
    };

    if (currentView !== 'landing' && currentView !== 'login') {
      loadData();
    }
  }, [currentView, selectedCity, user]);

  // Navigation component
  const Navigation = () => (
    <nav className="bg-gradient-to-r from-orange-600 via-red-600 to-orange-600 text-white shadow-2xl sticky top-0 z-50">
      <div className="max-w-7xl mx-auto px-4 py-4">
        <div className="flex justify-between items-center">
          <div 
            className="flex items-center space-x-3 cursor-pointer group"
            onClick={() => setCurrentView('landing')}
          >
            <div className="w-12 h-12 bg-white rounded-full flex items-center justify-center group-hover:scale-110 transition-transform">
              <span className="text-2xl">🏀</span>
            </div>
            <div>
              <h1 className="text-2xl font-bold">ANB Rising Stars</h1>
              <p className="text-xs opacity-90">Showcase 2025</p>
            </div>
          </div>
          
          <div className="flex items-center space-x-4 md:space-x-6">
            {user && (
              <>
                <button
                  onClick={() => setCurrentView('dashboard')}
                  className="hover:text-orange-200 transition-colors flex items-center space-x-1"
                >
                  <Home size={20} />
                  <span className="hidden md:inline">Inicio</span>
                </button>
                <button
                  onClick={() => setCurrentView('upload')}
                  className="hover:text-orange-200 transition-colors flex items-center space-x-1"
                >
                  <Upload size={20} />
                  <span className="hidden md:inline">Subir</span>
                </button>
                <button
                  onClick={() => setCurrentView('videos')}
                  className="hover:text-orange-200 transition-colors flex items-center space-x-1"
                >
                  <Video size={20} />
                  <span className="hidden md:inline">Videos</span>
                </button>
                <button
                  onClick={() => setCurrentView('rankings')}
                  className="hover:text-orange-200 transition-colors flex items-center space-x-1"
                >
                  <Trophy size={20} />
                  <span className="hidden md:inline">Rankings</span>
                </button>
              </>
            )}
            
            {user ? (
              <div className="flex items-center space-x-3">
                <button
                  onClick={() => setCurrentView('profile')}
                  className="flex items-center space-x-2 hover:text-orange-200"
                >
                  <div className="w-8 h-8 bg-white/20 rounded-full flex items-center justify-center">
                    <User size={16} />
                  </div>
                  <span className="hidden md:inline text-sm">{user.first_name}</span>
                </button>
                <button
                  onClick={async () => {
                    await apiService.logout();
                    setUser(null);
                    setCurrentView('landing');
                  }}
                  className="bg-white/20 backdrop-blur p-2 rounded-full hover:bg-white/30 transition-colors"
                >
                  <LogOut size={18} />
                </button>
              </div>
            ) : (
              <button
                onClick={() => setCurrentView('login')}
                className="bg-white text-orange-600 px-6 py-2 rounded-full font-bold hover:bg-orange-50 transition-all transform hover:scale-105 shadow-lg"
              >
                Iniciar Sesión
              </button>
            )}
          </div>
        </div>
      </div>
    </nav>
  );

  // Landing Page
  const LandingPage = () => {
    const [activeFeature, setActiveFeature] = useState(0);
    const features = [
      { icon: Upload, title: 'Sube tu Video', desc: 'Muestra tus mejores jugadas en 30 segundos' },
      { icon: Users, title: 'Votación Pública', desc: 'El público decide quiénes son los mejores' },
      { icon: Award, title: 'Clasificación', desc: 'Los más votados de cada ciudad clasifican' },
      { icon: Star, title: 'Showcase Final', desc: 'Compite frente a cazatalentos profesionales' }
    ];

    return (
      <div className="min-h-screen bg-gradient-to-br from-orange-600 to-orange-400">
        
        <div className="relative">
          <div className="absolute inset-0 bg-black/40"></div>
          
          <div className="relative max-w-7xl mx-auto px-4 py-16 md:py-24">
            <div className="text-center text-white">
              <div className="mb-8">
                <h1 className="text-5xl md:text-7xl lg:text-8xl font-black mb-4 animate-pulse">
                  RISING STARS
                </h1>
                <div className="text-3xl md:text-5xl lg:text-6xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-orange-400 to-yellow-400">
                  SHOWCASE 2025
                </div>
              </div>
              
              <p className="text-lg md:text-xl lg:text-2xl mb-12 opacity-90 max-w-3xl mx-auto leading-relaxed">
                ¿Tienes lo que se necesita para ser la próxima estrella del baloncesto nacional? 
                Demuestra tu talento y compite por un lugar en el torneo más importante del año.
              </p>
              
              <div className="flex flex-col sm:flex-row gap-4 justify-center mb-16">
                <button
                  onClick={() => setCurrentView('login')}
                  className="group bg-gradient-to-r from-orange-500 to-red-500 text-white px-8 py-4 rounded-full text-lg font-bold shadow-2xl transform transition-all duration-300 hover:scale-110 hover:shadow-orange-500/50"
                >
                  <Play className="inline mr-2 group-hover:animate-pulse" />
                  Comenzar Ahora
                </button>
                <button
                  onClick={() => setCurrentView('rankings')}
                  className="group bg-white text-gray-900 px-8 py-4 rounded-full text-lg font-bold shadow-2xl transform transition-all duration-300 hover:scale-110"
                >
                  <Trophy className="inline mr-2 group-hover:animate-bounce" />
                  Ver Rankings
                </button>
              </div>

              <div className="grid md:grid-cols-4 gap-6 mt-16">
                {features.map((feature, index) => {
                  const Icon = feature.icon;
                  return (
                    <div
                      key={index}
                      onMouseEnter={() => setActiveFeature(index)}
                      className={`bg-white/10 backdrop-blur-lg rounded-2xl p-6 transform transition-all duration-500 cursor-pointer ${
                        activeFeature === index ? 'scale-110 bg-white/20' : 'hover:scale-105'
                      }`}
                    >
                      <Icon className="w-12 h-12 text-orange-400 mx-auto mb-4" />
                      <h3 className="text-xl font-bold mb-2">{feature.title}</h3>
                      <p className="opacity-80 text-sm">{feature.desc}</p>
                    </div>
                  );
                })}
              </div>

              <div className="mt-16 p-8 bg-white/10 backdrop-blur-lg rounded-3xl">
                <h2 className="text-3xl font-bold mb-6">🏆 Premios y Beneficios</h2>
                <div className="grid md:grid-cols-3 gap-6 text-left">
                  <div className="bg-white/10 rounded-xl p-4">
                    <h3 className="font-bold text-lg mb-2 text-orange-400">Exposición Nacional</h3>
                    <p className="text-sm opacity-80">Muéstrate ante cazatalentos de equipos profesionales</p>
                  </div>
                  <div className="bg-white/10 rounded-xl p-4">
                    <h3 className="font-bold text-lg mb-2 text-orange-400">Entrenamiento Elite</h3>
                    <p className="text-sm opacity-80">Acceso a sesiones con entrenadores profesionales</p>
                  </div>
                  <div className="bg-white/10 rounded-xl p-4">
                    <h3 className="font-bold text-lg mb-2 text-orange-400">Contratos Profesionales</h3>
                    <p className="text-sm opacity-80">Oportunidad de firmar con equipos de la liga</p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  };

  // Login/Registration component  
  const LoginView = () => {
    const [isLogin, setIsLogin] = useState(true);
    const [formLoading, setFormLoading] = useState(false);
    const [error, setError] = useState('');
    const [formData, setFormData] = useState({
      email: '',
      password: '',
      firstName: '',
      lastName: '',
      city: '',
      country: 'Colombia',
      confirmPassword: ''
    });

    const handleSubmit = async () => {
      setFormLoading(true);
      setError('');

      try {
        if (isLogin) {
          await apiService.login(formData.email, formData.password);
          apiService.token = localStorage.getItem('access_token');
          const profile = await apiService.getProfile();
          setUser(profile);
          setCurrentView('dashboard');
        } else {
          if (formData.password !== formData.confirmPassword) {
            setError('Passwords do not match');
            return;
          }
          
          await apiService.signup(formData);
          await apiService.login(formData.email, formData.password);
          const profile = await apiService.getProfile();
          setUser(profile);
          setCurrentView('dashboard');
        }
      } catch (err) {
        setError(err.message || 'An error occurred');
      } finally {
        setFormLoading(false);
      }
    };

    return (
      <div className="min-h-screen bg-gradient-to-br from-gray-900 to-orange-900 flex items-center justify-center p-4">
        <div className="bg-white rounded-3xl shadow-2xl w-full max-w-md overflow-hidden">
          <div className="bg-gradient-to-r from-orange-500 to-red-500 p-6 text-white">
            <h2 className="text-3xl font-bold text-center">
              {isLogin ? 'Bienvenido de vuelta' : 'Únete a Rising Stars'}
            </h2>
            <p className="text-center mt-2 opacity-90">
              {isLogin ? 'Ingresa para ver tu progreso' : 'Comienza tu camino al estrellato'}
            </p>
          </div>
          
          <div className="p-8 space-y-4">
            {error && (
              <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-xl">
                {error}
              </div>
            )}

            {!isLogin && (
              <>
                <div className="grid grid-cols-2 gap-3">
                  <input
                    type="text"
                    placeholder="Nombre *"
                    required={!isLogin}
                    className="p-3 border-2 rounded-xl focus:ring-2 focus:ring-orange-500 focus:border-transparent transition-all"
                    value={formData.firstName}
                    onChange={(e) => setFormData({...formData, firstName: e.target.value})}
                  />
                  <input
                    type="text"
                    placeholder="Apellido *"
                    required={!isLogin}
                    className="p-3 border-2 rounded-xl focus:ring-2 focus:ring-orange-500 focus:border-transparent transition-all"
                    value={formData.lastName}
                    onChange={(e) => setFormData({...formData, lastName: e.target.value})}
                  />
                </div>
                
                <select
                  className="w-full p-3 border-2 rounded-xl focus:ring-2 focus:ring-orange-500 transition-all"
                  value={formData.city}
                  required={!isLogin}
                  onChange={(e) => setFormData({...formData, city: e.target.value})}
                >
                  <option value="">Selecciona tu ciudad *</option>
                  {cities.slice(1).map(city => (
                    <option key={city} value={city}>{city}</option>
                  ))}
                </select>
              </>
            )}
            
            <input
              type="email"
              placeholder="Correo electrónico *"
              required
              className="w-full p-3 border-2 rounded-xl focus:ring-2 focus:ring-orange-500 focus:border-transparent transition-all"
              value={formData.email}
              onChange={(e) => setFormData({...formData, email: e.target.value})}
            />
            
            <input
              type="password"
              placeholder="Contraseña *"
              required
              className="w-full p-3 border-2 rounded-xl focus:ring-2 focus:ring-orange-500 focus:border-transparent transition-all"
              value={formData.password}
              onChange={(e) => setFormData({...formData, password: e.target.value})}
            />

            {!isLogin && (
              <>
                <input
                  type="password"
                  placeholder="Confirmar contraseña *"
                  required={!isLogin}
                  className="w-full p-3 border-2 rounded-xl focus:ring-2 focus:ring-orange-500 focus:border-transparent transition-all"
                  value={formData.confirmPassword}
                  onChange={(e) => setFormData({...formData, confirmPassword: e.target.value})}
                />

                <div className="flex items-start space-x-2 text-sm text-gray-600">
                  <input type="checkbox" className="mt-1" />
                  <p>Acepto los términos y condiciones y autorizo el uso de mi imagen para fines promocionales del torneo</p>
                </div>
              </>
            )}
            
            <button
              onClick={handleSubmit}
              disabled={formLoading}
              className="w-full bg-gradient-to-r from-orange-500 to-red-500 text-white py-3 rounded-xl font-bold shadow-lg transform transition-all hover:scale-105 hover:shadow-xl disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {formLoading ? 'Cargando...' : (isLogin ? 'Ingresar' : 'Crear Cuenta')}
            </button>
          </div>
          
          <div className="pb-6 text-center">
            <p className="text-gray-600">
              {isLogin ? '¿No tienes cuenta?' : '¿Ya tienes cuenta?'}
              <button
                onClick={() => setIsLogin(!isLogin)}
                className="text-orange-600 font-bold ml-2 hover:underline"
              >
                {isLogin ? 'Regístrate' : 'Inicia Sesión'}
              </button>
            </p>
          </div>
        </div>
      </div>
    );
  };

  // Dashboard mejorado
  const Dashboard = () => {
    const totalVotes = myVideos.reduce((sum, video) => sum + (video.votes || 0), 0);
    const processedVideos = myVideos.filter(v => v.status === 'processed').length;
    const userRanking = rankings.findIndex(r => r.username === `${user?.first_name} ${user?.last_name}`) + 1;

    const stats = [
      { label: 'Votos Totales', value: totalVotes.toLocaleString(), change: '+' + Math.floor(totalVotes * 0.15), icon: ThumbsUp, color: 'from-orange-500 to-red-500' },
      { label: 'Ranking Ciudad', value: userRanking > 0 ? `#${userRanking}` : '-', change: userRanking > 0 ? '↑ 3' : '', icon: Trophy, color: 'from-purple-500 to-pink-500' },
      { label: 'Videos Procesados', value: processedVideos, change: `+${myVideos.length - processedVideos} pendientes`, icon: Video, color: 'from-blue-500 to-cyan-500' },
      { label: 'Días Restantes', value: '14', change: '', icon: Calendar, color: 'from-green-500 to-emerald-500' }
    ];

    return (
    <div className="min-h-screen bg-gray-50 p-4">
      <div className="max-w-7xl mx-auto">
        <div className="mb-8">
          <h1 className="text-4xl font-bold text-gray-800 mb-2">
            Hola, {user?.first_name} 👋
          </h1>
          <p className="text-gray-600">Este es tu panel de control para Rising Stars Showcase 2025</p>
          </div>
          
          <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
            {stats.map((stat, index) => {
              const Icon = stat.icon;
              return (
                <div
                  key={index}
                  className="bg-white rounded-2xl shadow-lg overflow-hidden transform hover:scale-105 transition-all"
                >
                  <div className={`h-2 bg-gradient-to-r ${stat.color}`}></div>
                  <div className="p-6">
                    <div className="flex items-start justify-between mb-4">
                      <Icon className="w-8 h-8 text-gray-400" />
                      {stat.change && (
                        <span className={`text-sm font-bold ${stat.change.includes('+') || stat.change.includes('↑') ? 'text-green-500' : 'text-gray-500'}`}>
                          {stat.change}
                        </span>
                      )}
                    </div>
                    <div className="text-3xl font-bold text-gray-800 mb-1">{stat.value}</div>
                    <div className="text-sm text-gray-600">{stat.label}</div>
                  </div>
                </div>
              );
            })}
          </div>

          <div className="grid lg:grid-cols-3 gap-6">
            <div className="lg:col-span-2 bg-white rounded-2xl shadow-lg p-6">
              <h3 className="text-xl font-bold mb-4 text-gray-800 flex items-center">
                <Video className="mr-2" />
                Mis Videos de Competencia
              </h3>
              
              {myVideos.length > 0 ? (
                <div className="space-y-4">
                  {myVideos.slice(0, 3).map(video => (
                    <div key={video.video_id} className="bg-gradient-to-br from-gray-900 to-gray-700 rounded-xl p-4 relative overflow-hidden group">
                      <div className="absolute inset-0 bg-gradient-to-t from-black/60 to-transparent"></div>
                      <div className="relative z-10 text-white">
                        <div className="flex items-center justify-between mb-2">
                          <h4 className="text-lg font-semibold truncate">{video.title}</h4>
                          <span className={`px-3 py-1 rounded-full text-xs font-bold ${
                            video.status === 'processed' ? 'bg-green-500' :
                            video.status === 'processing' ? 'bg-yellow-500' :
                            video.status === 'uploaded' ? 'bg-blue-500' : 'bg-red-500'
                          }`}>
                            {video.status === 'processed' ? 'APROBADO' :
                             video.status === 'processing' ? 'PROCESANDO' :
                             video.status === 'uploaded' ? 'SUBIDO' : 'ERROR'}
                          </span>
                        </div>
                        <div className="flex items-center justify-between text-sm">
                          <span>Votos: {video.votes || 0}</span>
                          <span>Subido: {new Date(video.uploaded_at).toLocaleDateString()}</span>
                        </div>
                      </div>
                    </div>
                  ))}
                  
                  {myVideos.length > 3 && (
                    <button
                      onClick={() => setCurrentView('profile')}
                      className="w-full text-center py-3 text-orange-600 hover:text-orange-700 font-semibold"
                    >
                      Ver todos mis videos ({myVideos.length})
                    </button>
                  )}
                </div>
              ) : (
                <div className="text-center py-8">
                  <Upload className="w-16 h-16 mx-auto text-gray-300 mb-4" />
                  <h4 className="text-lg font-semibold text-gray-600 mb-2">¡Sube tu primer video!</h4>
                  <p className="text-gray-500 mb-4">Muestra tus mejores jugadas y comienza a competir</p>
                  <button
                    onClick={() => setCurrentView('upload')}
                    className="bg-gradient-to-r from-orange-500 to-red-500 text-white px-6 py-3 rounded-full font-bold hover:shadow-lg transform hover:scale-105 transition-all"
                  >
                    Subir Video
                  </button>
                </div>
              )}
        </div>
        
        <div className="bg-white rounded-2xl shadow-lg p-6">
              <h3 className="text-xl font-bold mb-4 text-gray-800 flex items-center">
                <BarChart3 className="mr-2" />
                Tu Progreso
              </h3>
              <div className="space-y-4">
                <div>
                  <div className="flex justify-between text-sm mb-1">
                    <span className="text-gray-600">Objetivo de votos</span>
                    <span className="font-bold">{totalVotes} / 3,000</span>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-3 overflow-hidden">
                    <div className="bg-gradient-to-r from-orange-500 to-red-500 h-full rounded-full transition-all" 
                         style={{ width: `${Math.min((totalVotes / 3000) * 100, 100)}%` }}></div>
                  </div>
                </div>
                
                <div className="pt-4 border-t">
                  <p className="text-sm text-gray-600 mb-3">Posición en tu ciudad</p>
                  <div className="flex items-center justify-between">
                    <div className="text-2xl font-bold text-orange-600">
                      {userRanking > 0 ? `#${userRanking}` : '-'}
                    </div>
                    <div className="text-sm text-gray-500">de {rankings.length} participantes</div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <div className="mt-8 bg-white rounded-2xl shadow-lg p-6">
            <div className="flex items-center justify-between mb-6">
              <h3 className="text-xl font-bold text-gray-800">Videos Destacados para Votar</h3>
              <button
                onClick={() => setCurrentView('videos')}
                className="text-orange-600 hover:text-orange-700 font-semibold flex items-center"
              >
                Ver todos
                <ChevronRight className="ml-1" size={20} />
              </button>
            </div>
            <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
              {videos.slice(0, 3).map(video => (
                <VideoCard key={video.video_id} video={video} />
              ))}
            </div>
        </div>
      </div>
    </div>
  );
  };

  const UploadVideo = () => {
    const [selectedFile, setSelectedFile] = useState(null);
    const [title, setTitle] = useState('');
    const [isPublic, setIsPublic] = useState(true); // New state for public/private toggle
    const [uploading, setUploading] = useState(false);
    const [dragActive, setDragActive] = useState(false);
    const [error, setError] = useState('');
    const [errorDetails, setErrorDetails] = useState(null);
    const [processingStatus, setProcessingStatus] = useState(null);
    const [uploadProgress, setUploadProgress] = useState(0);
    const [errorModal, setErrorModal] = useState({ isOpen: false, title: '', message: '', technical: '', suggestions: [] });

    const handleDrag = (e) => {
      e?.preventDefault?.();
      e?.stopPropagation?.();
      if (e?.type === "dragenter" || e?.type === "dragover") {
        setDragActive(true);
      } else if (e?.type === "dragleave") {
        setDragActive(false);
      }
    };

    const handleDrop = (e) => {
      e?.preventDefault?.();
      e?.stopPropagation?.();
      setDragActive(false);
      if (e?.dataTransfer?.files?.[0]) {
        handleFileSelect(e.dataTransfer.files[0]);
      }
    };

    const handleFileSelect = (file) => {
      if (!file || !file.type.startsWith('video/')) {
        setError('Por favor selecciona un archivo de video válido');
        setErrorDetails(null);
        return;
      }
      
      // Check file size (100MB based on your requirements display)
      if (file.size > 100 * 1024 * 1024) {
        setError('El archivo es demasiado grande. Máximo 100MB.');
        setErrorDetails(null);
        return;
      }
      
      setSelectedFile(file);
      setError('');
      setErrorDetails(null);
      
      console.log('File details:', {
        name: file.name,
        size: file.size,
        type: file.type,
        sizeMB: (file.size / (1024 * 1024)).toFixed(2)
      });
    };

    const handleUpload = async () => {
    if (!selectedFile || !title.trim()) {
      setErrorModal({
        isOpen: true,
        title: 'Datos incompletos',
        message: 'Por favor proporciona un título y selecciona un archivo de video',
        technical: '',
        suggestions: ['Ingresa un título descriptivo', 'Selecciona un archivo de video válido']
      });
      return;
    }

    console.log('Starting upload with file details:', {
      name: selectedFile.name,
      size: selectedFile.size,
      sizeMB: (selectedFile.size / (1024 * 1024)).toFixed(2),
      type: selectedFile.type,
      isPublic: isPublic, // Include the public/private setting
      videoType: isPublic ? 'competencia' : 'prueba'
    });

    // Reset all states
    setUploading(true);
    setProcessingStatus('uploading');
    setUploadProgress(0);
    setError('');
    setErrorDetails(null);
    setErrorModal({ isOpen: false, title: '', message: '', technical: '', suggestions: [] });
    
    // Simulate upload progress
    let progress = 0;
    let interval = setInterval(() => {
      progress += Math.random() * 15;
      if (progress > 90) progress = 90;
      setUploadProgress(Math.round(progress));
    }, 500);

    try {
      // Pass the isPublic flag to your API service
      await apiService.uploadVideo(title.trim(), selectedFile, isPublic);
      
      // Success - clear interval and update states
      clearInterval(interval);
      interval = null;
      setUploadProgress(100);
      setProcessingStatus('processing');
      
      // Success timeout - but check state first
      setTimeout(() => {
        setProcessingStatus(current => {
          if (current === 'processing') {
            setUploading(false);
            return 'completed';
          }
          return current;
        });
      }, 3000);
      
    } catch (err) {
      // Error handling
      if (interval) {
        clearInterval(interval);
        interval = null;
      }
      
      console.error('Upload failed:', err);

      // Reset upload states immediately
      setUploading(false);
      setProcessingStatus('error');
      setUploadProgress(0);

      // Parse error message
      let errorMessage = 'Error desconocido al subir el video';
      if (err?.message) {
        errorMessage = err.message;
      } else if (err?.response?.data?.error) {
        errorMessage = err.response.data.error;
      } else if (err?.response?.data?.message) {
        errorMessage = err.response.data.message;
      } else if (typeof err === 'string') {
        errorMessage = err;
      }

      const errorInfo = parseBackendError(errorMessage);

      // Show error modal - this will stay visible until user dismisses it
      setErrorModal({
        isOpen: true,
        title: errorInfo.title,
        message: errorInfo.message,
        technical: errorInfo.technical,
        suggestions: errorInfo.suggestions
      });
    }
  };

    // Toggle Component
    const PublicPrivateToggle = () => {
      return (
        <div className="flex flex-col">
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Tipo de video
          </label>
          <div className="relative">
            <div className="flex bg-gray-100 rounded-full p-1 w-fit">
              <button
                type="button"
                onClick={() => setIsPublic(true)}
                className={`px-4 py-2 text-sm font-semibold rounded-full transition-all duration-200 ${
                  isPublic
                    ? 'bg-gradient-to-r from-orange-500 to-red-500 text-white shadow-md'
                    : 'text-gray-600 hover:text-gray-800'
                }`}
              >
                🏆 Competencia
              </button>
              <button
                type="button"
                onClick={() => setIsPublic(false)}
                className={`px-4 py-2 text-sm font-semibold rounded-full transition-all duration-200 ${
                  !isPublic
                    ? 'bg-gray-600 text-white shadow-md'
                    : 'text-gray-600 hover:text-gray-800'
                }`}
              >
                🔒 Prueba
              </button>
            </div>
            <p className="text-xs text-gray-500 mt-1">
              {isPublic 
                ? 'Visible públicamente y elegible para ranking' 
                : 'Solo visible para ti, no aparece en rankings'
              }
            </p>
          </div>
        </div>
      );
    };

    // Enhanced error display component
    const ErrorDisplay = ({ errorDetails }) => {
      if (!errorDetails) return null;

      return (
        <div className="bg-red-50 border border-red-200 rounded-xl p-6 mb-6">
          <div className="flex items-start space-x-3">
            <div className="flex-shrink-0">
              <XCircle className="w-6 h-6 text-red-500" />
            </div>
            <div className="flex-1">
              <h4 className="text-lg font-semibold text-red-800 mb-2">{errorDetails.title}</h4>
              {errorDetails.message && (
                <p className="text-red-700 mb-3">{errorDetails.message}</p>
              )}
              
              {errorDetails.suggestions.length > 0 && (
                <div className="mb-3">
                  <p className="text-sm font-medium text-red-800 mb-1">Soluciones sugeridas:</p>
                  <ul className="text-sm text-red-700 space-y-1">
                    {errorDetails.suggestions.map((suggestion, index) => (
                      <li key={index} className="flex items-start space-x-2">
                        <span className="text-red-500 mt-1">•</span>
                        <span>{suggestion}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              )}
              
              {errorDetails.technical && (
                <details className="mt-3">
                  <summary className="text-sm font-medium text-red-600 cursor-pointer hover:text-red-800">
                    Detalles técnicos
                  </summary>
                  <pre className="text-xs text-red-600 mt-2 bg-red-100 p-2 rounded overflow-x-auto">
                    {errorDetails.technical}
                  </pre>
                </details>
              )}
              
              <button
                onClick={() => {
                  setErrorDetails(null);
                  setProcessingStatus(null);
                }}
                className="mt-4 bg-red-600 text-white px-4 py-2 rounded-lg text-sm hover:bg-red-700 transition-colors"
              >
                Intentar de nuevo
              </button>
            </div>
          </div>
        </div>
      );
    };

  console.log("render", {processingStatus, uploading, selectedFile, error, errorDetails, isPublic});

  const handleModalClose = () => {
    setErrorModal({ isOpen: false, title: '', message: '', technical: '', suggestions: [] });
  };

  const handleRetry = () => {
    setErrorModal({ isOpen: false, title: '', message: '', technical: '', suggestions: [] });
    setProcessingStatus(null);
    setUploadProgress(0);
    setUploading(false);
  };

  const handleSelectNewFile = () => {
    setErrorModal({ isOpen: false, title: '', message: '', technical: '', suggestions: [] });
    setProcessingStatus(null);
    setUploadProgress(0);
    setUploading(false);
    setSelectedFile(null);
    setTitle('');
  };

  const parseBackendError = (errorMessage) => {
    console.log('Parsing error message:', errorMessage);
    
    let details = {
      title: 'Error al subir el video',
      message: errorMessage || 'Error desconocido al subir el video',
      technical: errorMessage,
      suggestions: []
    };

    if (errorMessage) {
      // Resolution error
      if (errorMessage.includes('resolution') && (errorMessage.includes('below minimum') || errorMessage.includes('minimum'))) {
        const resolutionMatch = errorMessage.match(/resolution (\d+x\d+)/);
        
        details.title = 'Resolución muy baja';
        details.message = resolutionMatch 
          ? `Tu video tiene resolución ${resolutionMatch[1]}. Se requiere mínimo 1920x1080 (Full HD)`
          : 'La resolución de tu video es muy baja';
        details.suggestions = [
          'Graba tu video en calidad Full HD (1920x1080) o superior',
          'Verifica la configuración de tu cámara antes de grabar',
          'Usa la cámara trasera de tu teléfono para mejor calidad',
          'Si usas iPhone, graba en modo "4K" o "1080p HD"'
        ];
      }
      // Duration error
      else if (errorMessage.includes('duration') && errorMessage.includes('range')) {
        const durationMatch = errorMessage.match(/duration ([\d.]+) seconds/);
        const rangeMatch = errorMessage.match(/range ([\d.]+)-([\d.]+) seconds/);
        
        details.title = 'Duración incorrecta';
        if (durationMatch && rangeMatch) {
          const currentDuration = parseFloat(durationMatch[1]);
          const minDuration = parseFloat(rangeMatch[1]);
          const maxDuration = parseFloat(rangeMatch[2]);
          details.message = `Tu video dura ${currentDuration} segundos. Debe durar entre ${minDuration} y ${maxDuration} segundos`;
          
          if (currentDuration < minDuration) {
            details.suggestions = ['Graba un video más largo con más jugadas', 'Incluye más movimientos y técnicas'];
          } else {
            details.suggestions = ['Edita tu video para reducir la duración', 'Enfócate en tus mejores jugadas'];
          }
        }
      }
      // File extension error
      else if (errorMessage.includes('invalid file extension')) {
        const extensionMatch = errorMessage.match(/invalid file extension: (\.\w+)/);
        details.title = 'Formato no válido';
        details.message = extensionMatch 
          ? `Archivo ${extensionMatch[1]} no permitido. Solo se acepta MP4`
          : 'Solo se acepta formato MP4';
        details.suggestions = [
          'Convierte tu video a formato MP4',
          'Usa aplicaciones como VLC o convertidores online',
          'Graba directamente en formato MP4'
        ];
      }
      // File size error
      else if (errorMessage.includes('file size') && errorMessage.includes('exceeds')) {
        const sizeMatch = errorMessage.match(/file size ([\d.]+\w+) exceeds maximum ([\d.]+\w+)/);
        details.title = 'Archivo muy grande';
        details.message = sizeMatch 
          ? `Tamaño actual: ${sizeMatch[1]}. Máximo: ${sizeMatch[2]}`
          : 'El archivo supera el tamaño máximo permitido (100MB)';
        details.suggestions = [
          'Reduce la calidad de video a 1080p',
          'Acorta la duración del video',
          'Usa un compresor de video online'
        ];
      }
    }

    return details;
  };

  const ErrorModal = ({ errorModal, onClose, onRetry, onSelectNewFile }) => {
    if (!errorModal.isOpen) return null;
      return (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-2xl shadow-2xl max-w-lg w-full max-h-[90vh] overflow-y-auto">
            {/* Header */}
            <div className="bg-red-500 text-white p-6 rounded-t-2xl">
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-3">
                  <XCircle className="w-8 h-8" />
                  <h3 className="text-xl font-bold">{errorModal.title}</h3>
                </div>
                <button
                  onClick={onClose}
                  className="text-white hover:text-red-200 transition-colors"
                >
                  <XCircle className="w-6 h-6" />
                </button>
              </div>
            </div>

            {/* Body */}
            <div className="p-6">
              {/* Error Message */}
              <div className="mb-6">
                <p className="text-gray-800 text-lg mb-4">{errorModal.message}</p>
                
                {/* Suggestions */}
                {errorModal.suggestions.length > 0 && (
                  <div className="bg-blue-50 border border-blue-200 rounded-xl p-4">
                    <h4 className="font-semibold text-blue-800 mb-3 flex items-center">
                      💡 Cómo solucionarlo:
                    </h4>
                    <ul className="space-y-2">
                      {errorModal.suggestions.map((suggestion, index) => (
                        <li key={index} className="flex items-start space-x-2 text-blue-700">
                          <span className="text-blue-500 mt-1 font-bold">•</span>
                          <span className="text-sm">{suggestion}</span>
                        </li>
                      ))}
                    </ul>
                  </div>
                )}
              </div>

              {errorModal.technical && (
                <details className="mb-6">
                  <summary className="text-sm font-medium text-gray-600 cursor-pointer hover:text-gray-800 mb-2">
                    Ver detalles técnicos
                  </summary>
                  <div className="bg-gray-100 p-3 rounded-lg text-xs text-gray-700 font-mono overflow-x-auto">
                    {errorModal.technical}
                  </div>
                </details>
              )}

              {/* Action Buttons */}
              <div className="flex flex-col sm:flex-row gap-3">
                <button
                  onClick={onRetry}
                  className="flex-1 bg-orange-500 text-white py-3 px-4 rounded-xl font-semibold hover:bg-orange-600 transition-colors"
                >
                  Intentar de nuevo
                </button>
                <button
                  onClick={onSelectNewFile}
                  className="flex-1 bg-gray-500 text-white py-3 px-4 rounded-xl font-semibold hover:bg-gray-600 transition-colors"
                >
                  Seleccionar otro archivo
                </button>
              </div>
            </div>
          </div>
        </div>
      );
    };

    return (
      <div className="min-h-screen bg-gray-50 p-4">
        <div className="max-w-4xl mx-auto">
          <h2 className="text-4xl font-bold mb-8 text-gray-800">Sube tu Video</h2>
          
          <div className="bg-white rounded-2xl shadow-lg overflow-hidden">
            <div className="bg-gradient-to-r from-orange-500 to-red-500 p-6 text-white">
              <h3 className="text-2xl font-bold mb-2">Requisitos del Video</h3>
              <div className="grid md:grid-cols-2 gap-4 text-sm">
                <div className="flex items-start space-x-2">
                  <CheckCircle size={16} className="mt-0.5 flex-shrink-0" />
                  <span>Duración: 20-60 segundos</span>
                </div>
                <div className="flex items-start space-x-2">
                  <CheckCircle size={16} className="mt-0.5 flex-shrink-0" />
                  <span>Formato: MP4</span>
                </div>
                <div className="flex items-start space-x-2">
                  <CheckCircle size={16} className="mt-0.5 flex-shrink-0" />
                  <span>Resolución mínima: 1080p (1920x1080)</span>
                </div>
                <div className="flex items-start space-x-2">
                  <CheckCircle size={16} className="mt-0.5 flex-shrink-0" />
                  <span>Tamaño máximo: 100MB</span>
                </div>
              </div>
            </div>

            <div className="p-8">
              {(error || errorDetails) && (
                <div className="mb-6">
                  {errorDetails ? (
                    <ErrorDisplay errorDetails={errorDetails} />
                  ) : (
                    <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-xl">
                      {error}
                    </div>
                  )}
                </div>
              )}

              {/* Title Input and Toggle Row */}
              {(!processingStatus || processingStatus === 'error') && (
                <div className="mb-6 grid md:grid-cols-2 gap-6">
                  {/* Title Input */}
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Título del video *
                    </label>
                    <input
                      type="text"
                      placeholder="Ej: Mejores jugadas - Juan Pérez"
                      value={title}
                      onChange={(e) => setTitle(e.target.value)}
                      className="w-full p-3 border-2 rounded-xl focus:ring-2 focus:ring-orange-500 focus:border-transparent transition-all"
                      maxLength={100}
                    />
                    <p className="text-sm text-gray-500 mt-1">{title.length}/100 caracteres</p>
                  </div>

                  {/* Public/Private Toggle */}
                  <PublicPrivateToggle />
                </div>
              )}

              {/* File Drop Zone */}
              {!selectedFile && (!processingStatus || processingStatus === 'error') && (
                <div
                  className={`border-3 border-dashed rounded-2xl p-12 text-center transition-all ${
                    dragActive ? 'border-orange-500 bg-orange-50' : 'border-gray-300 hover:border-orange-400'
                  }`}
                  onDragEnter={handleDrag}
                  onDragLeave={handleDrag}
                  onDragOver={handleDrag}
                  onDrop={handleDrop}
                >
                  <Upload className="w-20 h-20 mx-auto mb-4 text-gray-400" />
                  <p className="text-2xl mb-2 text-gray-700">Arrastra tu video aquí</p>
                  <p className="text-gray-500 mb-4">o</p>
                  <label className="cursor-pointer">
                    <input
                      type="file"
                      accept="video/*"
                      onChange={(e) => handleFileSelect(e.target.files[0])}
                      className="hidden"
                    />
                    <span className="bg-gradient-to-r from-orange-500 to-red-500 text-white px-8 py-3 rounded-full font-bold hover:shadow-lg transition-all inline-block">
                      Seleccionar archivo
                    </span>
                  </label>
                </div>
              )}

              {/* File Selected */}
              {selectedFile && !uploading && (!processingStatus || processingStatus === 'error') && (
                <div className="text-center">
                  <Video className="w-20 h-20 mx-auto mb-4 text-orange-500" />
                  <p className="text-xl mb-2 text-gray-700 font-semibold">{selectedFile.name}</p>
                  <div className="mb-6 space-y-2">
                    <p className="text-gray-500">
                      Tamaño: {(selectedFile.size / (1024 * 1024)).toFixed(2)} MB
                    </p>
                    <div className={`inline-flex items-center px-3 py-1 rounded-full text-sm font-medium ${
                      isPublic 
                        ? 'bg-orange-100 text-orange-800' 
                        : 'bg-gray-100 text-gray-800'
                    }`}>
                      {isPublic ? '🏆 Competencia' : '🔒 Prueba'} 
                      <span className="ml-2">
                        {isPublic ? '(Público)' : '(Privado)'}
                      </span>
                    </div>
                  </div>
                  <div className="flex gap-4 justify-center">
                    <button
                      onClick={() => {
                        setSelectedFile(null);
                        setError('');
                        setErrorDetails(null);
                        setProcessingStatus(null);
                      }}
                      className="bg-gray-200 text-gray-700 px-6 py-3 rounded-full font-semibold hover:bg-gray-300 transition-colors"
                    >
                      Cambiar archivo
                    </button>
                    <button
                      onClick={handleUpload}
                      className="bg-gradient-to-r from-orange-500 to-red-500 text-white px-8 py-3 rounded-full font-bold hover:shadow-lg transform hover:scale-105 transition-all"
                    >
                      Subir Video
                    </button>
                  </div>
                </div>
              )}

              {/* Processing Status */}
              {processingStatus && processingStatus !== 'error' && (
                <div className="space-y-6">
                  {processingStatus === 'uploading' && (
                    <div>
                      <div className="flex items-center justify-between mb-3">
                        <span className="text-lg font-semibold text-gray-700">Subiendo video...</span>
                        <span className="text-2xl font-bold text-orange-600">{uploadProgress}%</span>
                      </div>
                      <div className="w-full bg-gray-200 rounded-full h-4 overflow-hidden">
                        <div 
                          className="bg-gradient-to-r from-orange-500 to-red-500 h-full rounded-full transition-all duration-300"
                          style={{ width: `${uploadProgress}%` }}
                        />
                      </div>
                      <p className="text-sm text-gray-500 mt-2">No cierres esta ventana...</p>
                    </div>
                  )}

                  {processingStatus === 'processing' && (
                    <div className="text-center">
                      <div className="relative">
                        <Loader2 className="w-20 h-20 mx-auto mb-4 text-orange-500 animate-spin" />
                        <div className="absolute inset-0 flex items-center justify-center">
                          <span className="text-4xl">🏀</span>
                        </div>
                      </div>
                      <p className="text-2xl font-semibold text-gray-700 mb-2">Procesando tu video...</p>
                      <div className="space-y-2 text-gray-500">
                        <p className="flex items-center justify-center">
                          <CheckCircle className="text-green-500 mr-2" size={16} />
                          Validando formato y duración
                        </p>
                        <p className="flex items-center justify-center">
                          <Loader2 className="animate-spin mr-2" size={16} />
                          Optimizando calidad y resolución
                        </p>
                        <p className="flex items-center justify-center text-gray-400">
                          <Clock className="mr-2" size={16} />
                          Preparando para votación
                        </p>
                      </div>
                    </div>
                  )}

                  {processingStatus === 'completed' && (
                    <div className="text-center">
                      <CheckCircle className="w-20 h-20 mx-auto mb-4 text-green-500" />
                      <p className="text-3xl font-bold text-gray-800 mb-2">¡Video procesado con éxito!</p>
                      <p className="text-gray-600 mb-8">
                        {isPublic 
                          ? 'Tu video ya está disponible para votación pública' 
                          : 'Tu video de prueba ha sido guardado como privado'
                        }
                      </p>
                      <div className="bg-green-50 border border-green-200 rounded-xl p-4 mb-6 max-w-md mx-auto">
                        <p className="text-green-800 text-sm">
                          <strong>Próximos pasos:</strong> 
                          {isPublic 
                            ? ' Comparte tu video en redes sociales para conseguir más votos'
                            : ' Puedes cambiar la visibilidad a público desde tu dashboard cuando estés listo'
                          }
                        </p>
                      </div>
                      <button
                        onClick={() => setCurrentView('dashboard')}
                        className="bg-gradient-to-r from-green-500 to-emerald-500 text-white px-8 py-3 rounded-full font-bold hover:shadow-lg transform hover:scale-105 transition-all"
                      >
                        Ver mi dashboard
                      </button>
                    </div>
                  )}
                </div>
              )}
            </div>
          </div>
        </div>
        <ErrorModal
          errorModal={errorModal}
          onClose={handleModalClose}
          onRetry={handleRetry}
          onSelectNewFile={handleSelectNewFile}
        />
      </div>
    );
  };

  const VideosView = () => {
    const [filterPosition, setFilterPosition] = useState('todas');

    return (
      <div className="min-h-screen bg-gray-50 p-4">
        <div className="max-w-7xl mx-auto">
          <h2 className="text-4xl font-bold mb-8 text-gray-800">Videos de Competencia</h2>
          
          <div className="mb-6 flex flex-wrap gap-2">
            <div className="flex items-center space-x-2 text-sm text-gray-600 mr-4">
              <Filter size={16} />
              <span>Filtrar por:</span>
            </div>
            {cities.map(pos => (
              <button
                key={pos}
                onClick={() => setFilterPosition(pos.toLowerCase())}
                className={`px-4 py-2 rounded-full font-semibold transition-all ${
                  filterPosition === pos.toLowerCase()
                    ? 'bg-orange-500 text-white'
                    : 'bg-white text-gray-700 hover:bg-orange-100'
                }`}
              >
                {pos}
              </button>
            ))}
          </div>

          {loading ? (
            <div className="flex items-center justify-center h-64">
              <Loader2 className="w-8 h-8 animate-spin text-orange-500" />
            </div>
          ) : (
            <div className="grid md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
              {videos.map(video => (
                <VideoCard key={video.video_id} video={video} detailed />
              ))}
            </div>
          )}

          {!loading && videos.length === 0 && (
            <div className="text-center py-12">
              <Video className="w-16 h-16 mx-auto text-gray-400 mb-4" />
              <h3 className="text-xl font-semibold text-gray-600 mb-2">No hay videos disponibles</h3>
              <p className="text-gray-500">¡Sé el primero en subir un video!</p>
            </div>
          )}
        </div>
      </div>
    );
  };

  const Rankings = () => {
    const [rankings, setRankings] = useState([]);
    const [loading, setLoading] = useState(true);
    const [refreshing, setRefreshing] = useState(false);
    const [selectedCity, setSelectedCity] = useState('todas');
    
    const cities = ['Todas', 'Bogotá', 'Medellín', 'Cali', 'Barranquilla', 'Cartagena', 'Bucaramanga', 'Pereira'];

    // Function to load rankings
    const loadRankings = async (city = 'todas', showLoading = true) => {
        try {
            if (showLoading) setLoading(true);
            
            console.log('Loading rankings for city:', city);
            const result = await apiService.getTopRankings(10, city);
            console.log('API Result:', result);
            
            // Map the player rankings data to the expected format
            const mappedRankings = (result.rankings || []).map(player => ({
                user_id: player.user_id,
                username: `${player.first_name} ${player.last_name}`.trim() || 'Usuario Anónimo',
                first_name: player.first_name,
                last_name: player.last_name,
                email: player.email,
                city: player.city || 'No especificada',
                country: player.country,
                votes: player.total_votes || 0,
                ranking: player.ranking,
                title: `Jugador de ${player.city || 'Colombia'}`, // Generate a title
                last_updated: player.last_updated
            }));
            
            console.log('Mapped rankings:', mappedRankings);
            setRankings(mappedRankings);
            
        } catch (error) {
            console.error('Failed to load rankings:', error);
            setRankings([]);
        } finally {
            if (showLoading) setLoading(false);
        }
    };

    // Function to refresh rankings
    const refreshRankings = async () => {
        try {
            setRefreshing(true);
            console.log('Refreshing rankings...');
            
            // Call the refresh endpoint
            await apiService.refreshRankings();
            console.log('Rankings refreshed successfully');
            
            // Reload the rankings
            await loadRankings(selectedCity, false);
            
        } catch (error) {
            console.error('Failed to refresh rankings:', error);
            // Still try to reload even if refresh failed
            await loadRankings(selectedCity, false);
        } finally {
            setRefreshing(false);
        }
    };

    // Load rankings when component mounts or city changes
    useEffect(() => {
        loadRankings(selectedCity);
    }, [selectedCity]);

    return (
        <div className="min-h-screen bg-gray-50 p-4">
            <div className="max-w-6xl mx-auto">
                <div className="mb-8">
                    <div className="flex justify-between items-center">
                        <div>
                            <h2 className="text-4xl font-bold text-gray-800 mb-2">Rankings Rising Stars 2025</h2>
                            <p className="text-gray-600">Los mejores jugadores de cada ciudad competirán en el Showcase final</p>
                        </div>
                        
                        {/* Refresh Button */}
                        <button
                            onClick={refreshRankings}
                            disabled={refreshing || loading}
                            className={`flex items-center space-x-2 px-4 py-2 rounded-lg font-semibold transition-all ${
                                refreshing 
                                    ? 'bg-gray-300 text-gray-600 cursor-not-allowed' 
                                    : 'bg-gradient-to-r from-orange-500 to-red-500 text-white hover:shadow-lg transform hover:scale-105'
                            }`}
                        >
                            <RefreshCw className={`w-4 h-4 ${refreshing ? 'animate-spin' : ''}`} />
                            <span>{refreshing ? 'Actualizando...' : 'Actualizar Rankings'}</span>
                        </button>
                    </div>
                </div>
                
                <div className="mb-6 flex flex-wrap gap-2">
                    {cities.map(city => (
                        <button
                            key={city}
                            onClick={() => setSelectedCity(city.toLowerCase())}
                            className={`px-5 py-2 rounded-full font-semibold transition-all ${
                                selectedCity === city.toLowerCase()
                                    ? 'bg-gradient-to-r from-orange-500 to-red-500 text-white shadow-lg'
                                    : 'bg-white text-gray-700 hover:bg-orange-50 shadow'
                            }`}
                        >
                            {city}
                        </button>
                    ))}
                </div>

                <div className="bg-white rounded-2xl shadow-xl overflow-hidden">
                    <div className="bg-gradient-to-r from-orange-500 via-red-500 to-orange-500 p-6 text-white">
                        <h3 className="text-2xl font-bold">
                            Top Jugadores {selectedCity !== 'todas' ? `- ${selectedCity.charAt(0).toUpperCase() + selectedCity.slice(1)}` : 'Nacional'}
                        </h3>
                        <p className="text-sm opacity-90 mt-1">Actualizado en tiempo real</p>
                    </div>
                    
                    <div className="divide-y divide-gray-100">
                        {loading ? (
                            <div className="flex items-center justify-center py-12">
                                <Loader2 className="w-8 h-8 animate-spin text-orange-500" />
                            </div>
                        ) : rankings.length > 0 ? (
                            rankings.map((player, index) => (
                                <div key={player.user_id} className="p-6 hover:bg-gray-50 transition-all group">
                                    <div className="flex items-center justify-between">
                                        <div className="flex items-center space-x-4">
                                            <div className={`text-3xl font-black ${index === 0 ? 'text-yellow-500' : index === 1 ? 'text-gray-400' : index === 2 ? 'text-orange-600' : 'text-gray-300'}`}>
                                                {index === 0 ? '🥇' : index === 1 ? '🥈' : index === 2 ? '🥉' : `#${index + 1}`}
                                            </div>
                                            <div className="w-16 h-16 bg-gradient-to-br from-orange-400 to-red-400 rounded-full flex items-center justify-center text-2xl shadow-lg group-hover:scale-110 transition-transform text-white font-bold">
                                                {player.username ? player.username.charAt(0).toUpperCase() : '?'}
                                            </div>
                                            <div>
                                                <h4 className="text-xl font-bold text-gray-800">{player.username}</h4>
                                                <div className="flex items-center space-x-4 text-sm text-gray-600 mt-1">
                                                    <span className="flex items-center">
                                                        <MapPin size={14} className="mr-1" />
                                                        {player.city}
                                                    </span>
                                                    <span className="bg-gray-100 px-2 py-0.5 rounded-full">{player.title}</span>
                                                </div>
                                            </div>
                                        </div>
                                        
                                        <div className="flex items-center space-x-6">
                                            <div className="text-right">
                                                <div className="text-3xl font-bold text-gray-800">{(player.votes || 0).toLocaleString()}</div>
                                                <div className="text-sm text-gray-500">votos</div>
                                            </div>
                                            
                                            <button className="bg-gradient-to-r from-orange-500 to-red-500 text-white px-6 py-2 rounded-full font-semibold hover:shadow-lg transform hover:scale-105 transition-all">
                                                Ver Perfil
                                            </button>
                                        </div>
                                    </div>
                                    
                                    {index < 3 && (
                                        <div className="mt-4 pt-4 border-t border-gray-100">
                                            <div className="flex items-center justify-between text-sm">
                                                <span className="text-gray-500">Clasificado para el Showcase Final</span>
                                                <span className="text-green-600 font-semibold flex items-center">
                                                    <CheckCircle size={16} className="mr-1" />
                                                    Confirmado
                                                </span>
                                            </div>
                                        </div>
                                    )}
                                </div>
                            ))
                        ) : (
                            <div className="p-12 text-center">
                                <Trophy className="w-16 h-16 mx-auto text-gray-400 mb-4" />
                                <h3 className="text-xl font-semibold text-gray-600 mb-2">No hay rankings disponibles</h3>
                                <p className="text-gray-500 mb-4">Los rankings aparecerán cuando haya jugadores con votos.</p>
                                <button
                                    onClick={refreshRankings}
                                    className="bg-gradient-to-r from-orange-500 to-red-500 text-white px-6 py-2 rounded-full font-semibold hover:shadow-lg transform hover:scale-105 transition-all"
                                >
                                    Actualizar Rankings
                                </button>
                            </div>
                        )}
                    </div>
                </div>
                
                <div className="mt-8 bg-gradient-to-r from-orange-100 to-red-100 rounded-2xl p-6">
                    <h3 className="text-lg font-bold text-gray-800 mb-2">📊 Estadísticas de Votación</h3>
                    <div className="grid md:grid-cols-4 gap-4 text-center">
                        <div>
                            <div className="text-2xl font-bold text-orange-600">{rankings.reduce((sum, r) => sum + (r.votes || 0), 0).toLocaleString()}</div>
                            <div className="text-sm text-gray-600">Votos totales</div>
                        </div>
                        <div>
                            <div className="text-2xl font-bold text-purple-600">{rankings.length}</div>
                            <div className="text-sm text-gray-600">Participantes</div>
                        </div>
                        <div>
                            <div className="text-2xl font-bold text-blue-600">{cities.length - 1}</div>
                            <div className="text-sm text-gray-600">Ciudades activas</div>
                        </div>
                        <div>
                            <div className="text-2xl font-bold text-green-600">14</div>
                            <div className="text-sm text-gray-600">Días restantes</div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    );
  };

  // Perfil mejorado
  const Profile = () => (
    <div className="min-h-screen bg-gray-50 p-4">
      <div className="max-w-4xl mx-auto">
        <div className="bg-white rounded-2xl shadow-xl overflow-hidden">
          <div className="bg-gradient-to-r from-orange-500 to-red-500 h-32"></div>
          <div className="px-8 pb-8">
            <div className="flex items-end -mt-16 mb-6">
              <div className="w-32 h-32 bg-white rounded-full border-4 border-white shadow-xl flex items-center justify-center text-5xl">
                🏀
              </div>
              <div className="ml-6 mb-4">
                <h2 className="text-3xl font-bold text-gray-800">{user?.first_name} {user?.last_name}</h2>
                <p className="text-gray-600">{user?.email}</p>
                <div className="flex items-center space-x-3 mt-2">
                  <span className="bg-orange-100 text-orange-700 px-3 py-1 rounded-full text-sm font-semibold">
                    {user?.city}
                  </span>
                  <span className="bg-blue-100 text-blue-700 px-3 py-1 rounded-full text-sm font-semibold">
                    {user?.country}
                  </span>
                  <span className="bg-green-100 text-green-700 px-3 py-1 rounded-full text-sm font-semibold flex items-center">
                    <CheckCircle size={14} className="mr-1" />
                    Verificado
                  </span>
                </div>
              </div>
            </div>

            <div className="grid md:grid-cols-4 gap-4 mb-8">
              <div className="bg-gradient-to-br from-orange-50 to-red-50 rounded-xl p-4 text-center">
                <div className="text-3xl font-bold text-orange-600">
                  {myVideos.reduce((sum, video) => sum + (video.votes || 0), 0)}
                </div>
                <div className="text-sm text-gray-600">Votos totales</div>
              </div>
              <div className="bg-gradient-to-br from-purple-50 to-pink-50 rounded-xl p-4 text-center">
                <div className="text-3xl font-bold text-purple-600">
                  #{rankings.findIndex(r => r.username === `${user?.first_name} ${user?.last_name}`) + 1 || '-'}
                </div>
                <div className="text-sm text-gray-600">Ranking ciudad</div>
              </div>
              <div className="bg-gradient-to-br from-blue-50 to-cyan-50 rounded-xl p-4 text-center">
                <div className="text-3xl font-bold text-blue-600">{myVideos.length}</div>
                <div className="text-sm text-gray-600">Videos subidos</div>
              </div>
              <div className="bg-gradient-to-br from-green-50 to-emerald-50 rounded-xl p-4 text-center">
                <div className="text-3xl font-bold text-green-600">
                  {myVideos.filter(v => v.status === 'processed').length}
                </div>
                <div className="text-sm text-gray-600">Videos procesados</div>
              </div>
            </div>

            <div className="space-y-6">
              <div className="border rounded-xl p-6">
                <h3 className="text-xl font-bold mb-4 text-gray-800 flex items-center">
                  <Video className="mr-2" />
                  Mis Videos de Competencia
                </h3>
                {myVideos.length > 0 ? (
                  <div className="space-y-3">
                    {myVideos.map(video => (
                      <div key={video.video_id} className="bg-gray-50 rounded-xl p-4">
                        <div className="flex items-center justify-between">
                          <div>
                            <h4 className="font-semibold text-gray-800">{video.title}</h4>
                            <p className="text-sm text-gray-600">
                              Subido: {new Date(video.uploaded_at).toLocaleDateString()}
                            </p>
                          </div>
                          <div className="flex items-center space-x-3">
                            <span className={`px-2 py-1 rounded-full text-xs font-medium ${
                              video.status === 'processed' ? 'bg-green-100 text-green-800' :
                              video.status === 'processing' ? 'bg-yellow-100 text-yellow-800' :
                              video.status === 'uploaded' ? 'bg-blue-100 text-blue-800' :
                              'bg-red-100 text-red-800'
                            }`}>
                              {video.status === 'processed' ? 'Procesado' :
                               video.status === 'processing' ? 'Procesando' :
                               video.status === 'uploaded' ? 'Subido' : 'Error'}
                            </span>
                            <span className="text-sm font-bold text-gray-700">{video.votes || 0} votos</span>
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                ) : (
                  <div className="text-center py-8">
                    <Upload className="w-16 h-16 mx-auto text-gray-300 mb-4" />
                    <h4 className="text-lg font-semibold text-gray-600 mb-2">¡Sube tu primer video!</h4>
                    <p className="text-gray-500 mb-4">Muestra tus mejores jugadas y comienza a competir</p>
                    <button
                      onClick={() => setCurrentView('upload')}
                      className="bg-gradient-to-r from-orange-500 to-red-500 text-white px-6 py-3 rounded-full font-bold hover:shadow-lg transform hover:scale-105 transition-all"
                    >
                      Subir Video
                    </button>
                  </div>
                )}
              </div>

              <div className="border rounded-xl p-6">
                <h3 className="text-xl font-bold mb-4 text-gray-800 flex items-center">
                  <Shield className="mr-2" />
                  Configuración de Privacidad
                </h3>
                <div className="space-y-3">
                  <label className="flex items-center justify-between">
                    <span className="text-gray-700">Perfil público</span>
                    <input type="checkbox" defaultChecked className="w-5 h-5 text-orange-500" />
                  </label>
                  <label className="flex items-center justify-between">
                    <span className="text-gray-700">Mostrar estadísticas</span>
                    <input type="checkbox" defaultChecked className="w-5 h-5 text-orange-500" />
                  </label>
                  <label className="flex items-center justify-between">
                    <span className="text-gray-700">Recibir notificaciones</span>
                    <input type="checkbox" defaultChecked className="w-5 h-5 text-orange-500" />
                  </label>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );

  const SimpleVideoView = ({ video, onBack, onVote, hasVoted }) => {
    return (
      <div className="min-h-screen bg-gray-50 p-4">
        <div className="max-w-4xl mx-auto">
          {/* Back Button */}
          <button
            onClick={onBack}
            className="flex items-center space-x-2 text-gray-600 hover:text-gray-800 mb-6"
          >
            <ChevronLeft className="w-5 h-5" />
            <span>Volver a videos</span>
          </button>

          {/* Video Player Card */}
          <div className="bg-white rounded-2xl shadow-xl overflow-hidden">
            {/* Video Container */}
            <div className="relative bg-black">
              <video
                className="w-full h-auto max-h-[70vh] object-contain"
                controls
                autoPlay
                poster="/api/placeholder/800/450"
              >
                <source 
                  src={video.video_url || `/api/videos/${video.video_id}/stream`} 
                  type="video/mp4" 
                />
                  Tu navegador no soporta el elemento de video.
                
              </video>
            </div>

            {/* Video Info */}
            <div className="p-6">
              <div className="grid md:grid-cols-3 gap-6">
                {/* Left side - Video details */}
                <div className="md:col-span-2">
                  <h1 className="text-3xl font-bold text-gray-800 mb-3">{video.title}</h1>
                  
                  <div className="flex items-center space-x-4 text-gray-600 mb-4">
                    <div className="flex items-center space-x-2">
                      <User className="w-5 h-5" />
                      <span className="font-medium text-lg">{video.user_first_name} {video.user_last_name}</span>
                    </div>
                    <div className="flex items-center space-x-2">
                      <MapPin className="w-5 h-5" />
                      <span>{video.user_city}</span>
                    </div>
                  </div>

                  <div className="flex items-center space-x-6 mb-6">
                    <div className="flex items-center space-x-2">
                      <ThumbsUp className="w-6 h-6 text-orange-500" />
                      <span className="font-bold text-2xl text-gray-800">{(video.votes || 0).toLocaleString()}</span>
                      <span className="text-gray-600">votos</span>
                    </div>
                    <div className="flex items-center space-x-2">
                      <Eye className="w-6 h-6 text-blue-500" />
                      <span className="font-bold text-2xl text-gray-800">{Math.floor(Math.random() * 5000 + 1000).toLocaleString()}</span>
                      <span className="text-gray-600">vistas</span>
                    </div>
                    <div className="flex items-center space-x-2">
                      <Calendar className="w-5 h-5 text-gray-400" />
                      <span className="text-gray-600">{new Date(video.uploaded_at).toLocaleDateString()}</span>
                    </div>
                  </div>
                </div>

                {/* Right side - Voting */}
                <div className="bg-gradient-to-br from-orange-50 to-red-50 rounded-xl p-6">
                  <h3 className="text-xl font-bold text-gray-800 mb-4 text-center">
                    ¿Te gustó este video?
                  </h3>
                  
                  <button
                    onClick={() => onVote(video.video_id)}
                    disabled={hasVoted || video.status !== 'processed'}
                    className={`w-full py-4 rounded-xl font-bold text-lg transition-all transform ${
                      hasVoted
                        ? 'bg-green-500 text-white'
                        : video.status !== 'processed'
                        ? 'bg-gray-300 text-gray-500 cursor-not-allowed'
                        : 'bg-gradient-to-r from-orange-500 to-red-500 text-white hover:shadow-xl hover:scale-105'
                    }`}
                  >
                    {hasVoted ? (
                      <>
                        <CheckCircle className="inline mr-2" />
                        ¡Votado!
                      </>
                    ) : (
                      <>
                        <ThumbsUp className="inline mr-2" />
                        Votar
                      </>
                    )}
                  </button>

                  {!hasVoted && video.status === 'processed' && (
                    <p className="text-sm text-gray-600 text-center mt-3">
                      Tu voto ayuda a este jugador a clasificar
                    </p>
                  )}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  };

  const VideoCard = ({ video, detailed = false }) => {
    const [voted, setVoted] = useState(votedVideos.has(video.video_id));
    
    // Simple click handler - just navigate to expanded view
    const handleVideoClick = () => {
      if (video.status === 'processed') {
        setExpandedVideo(video);
        setCurrentView('video-expanded');
      }
    };
    
    const handleVote = async (e) => {
      e.stopPropagation(); // Prevent opening video
      if (!voted && video.status === 'processed' && user) {
        try {
          await apiService.voteVideo(video.video_id);
          setVoted(true);
          setVotedVideos(new Set([...votedVideos, video.video_id]));
          
          const updatedVideos = await apiService.getPublicVideos();
          setVideos(Array.isArray(updatedVideos) ? updatedVideos : []);
        } catch (error) {
          console.error('Vote failed:', error);
        }
      }
    };

    return (
      <div className="bg-white rounded-xl shadow-lg overflow-hidden transform hover:scale-105 transition-all duration-300 group">
        <div 
          className="relative h-48 bg-gradient-to-br from-gray-800 to-gray-900 flex items-center justify-center cursor-pointer"
          onClick={handleVideoClick}
        >
          <Video className="w-12 h-12 text-white opacity-50" />
          
          {video.status === 'processing' && (
            <div className="absolute inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center">
              <div className="text-center">
                <Loader2 className="w-8 h-8 text-white animate-spin mx-auto mb-2" />
                <span className="text-white text-sm">Procesando...</span>
              </div>
            </div>
          )}
          
          {video.status === 'processed' && (
            <div className="absolute inset-0 bg-black/0 hover:bg-black/40 transition-all flex items-center justify-center opacity-0 hover:opacity-100">
              <button className="bg-orange-500 text-white px-6 py-3 rounded-full font-semibold transform scale-0 group-hover:scale-100 transition-transform flex items-center space-x-2">
                <Play className="w-5 h-5" />
                <span>Ver Video</span>
              </button>
            </div>
          )}
          
          <div className="absolute top-2 right-2 bg-black/50 backdrop-blur text-white px-2 py-1 rounded-full text-xs">
            <Eye className="inline mr-1" size={12} />
            {Math.floor(Math.random() * 5000 + 1000)}
          </div>

          <div className="absolute bottom-2 right-2 bg-black/70 text-white px-2 py-1 rounded text-xs">
            {Math.floor(Math.random() * 40 + 20)}s
          </div>
        </div>
        
        <div className="p-4">
          <h4 className="font-bold text-lg text-gray-800 mb-1 truncate">{video.title}</h4>
          <div className="flex items-center justify-between text-sm text-gray-600 mb-3">
            <span>{video.user_first_name} {video.user_last_name}</span>
            <span className="bg-gray-100 px-2 py-0.5 rounded-full text-xs flex items-center">
              <MapPin className="w-3 h-3 mr-1" />
              {video.user_city}
            </span>
          </div>
          
          <div className="flex items-center justify-between">
            <div>
              <span className="text-2xl font-bold text-gray-800">{(video.votes || 0).toLocaleString()}</span>
              <span className="text-sm text-gray-500 ml-1">votos</span>
            </div>
            <button
              onClick={handleVote}
              disabled={voted || video.status !== 'processed' || !user}
              className={`px-4 py-2 rounded-full font-semibold transition-all transform ${
                voted
                  ? 'bg-green-500 text-white'
                  : video.status === 'processing'
                  ? 'bg-gray-200 text-gray-400 cursor-not-allowed'
                  : !user
                  ? 'bg-gray-200 text-gray-500 cursor-not-allowed'
                  : 'bg-gradient-to-r from-orange-500 to-red-500 text-white hover:shadow-lg hover:scale-105'
              }`}
            >
              {voted ? (
                <>
                  <CheckCircle className="inline mr-1" size={16} />
                  Votado
                </>
              ) : video.status === 'processing' ? (
                'Procesando...'
              ) : !user ? (
                'Inicia sesión'
              ) : (
                <>
                  <ThumbsUp className="inline mr-1" size={16} />
                  Votar
                </>
              )}
            </button>
          </div>
        </div>
      </div>
    );
  };

  const handleVoteFromExpanded = async (videoId) => {
    if (!user || votedVideos.has(videoId)) return;
    
    try {
      await apiService.voteVideo(videoId);
      setVotedVideos(new Set([...votedVideos, videoId]));
      
      // Update expanded video vote count
      if (expandedVideo && expandedVideo.video_id === videoId) {
        setExpandedVideo({
          ...expandedVideo,
          votes: (expandedVideo.votes || 0) + 1
        });
      }
      
      const updatedVideos = await apiService.getPublicVideos();
      setVideos(Array.isArray(updatedVideos) ? updatedVideos : []);
    } catch (error) {
      console.error('Vote failed:', error);
    }
  };

  // Renderizado principal
  return (
    <div className="min-h-screen bg-gray-50">
      <Navigation />
      {currentView === 'landing' && <LandingPage />}
      {currentView === 'login' && <LoginView />}
      {currentView === 'dashboard' && user && <Dashboard />}
      {currentView === 'upload' && user && <UploadVideo />}
      {currentView === 'videos' && <VideosView />}
      {currentView === 'rankings' && <Rankings />}
      {currentView === 'profile' && user && <Profile />}
      {currentView === 'video-expanded' && expandedVideo && (
        <SimpleVideoView
          video={expandedVideo}
          onBack={() => setCurrentView('videos')}
          onVote={handleVoteFromExpanded}
          hasVoted={votedVideos.has(expandedVideo.video_id)}
        />
      )}
      {!user && currentView !== 'landing' && currentView !== 'login' && currentView !== 'rankings' && currentView !== 'videos' && (
        <div className="min-h-screen flex items-center justify-center">
          <div className="bg-white rounded-xl shadow-lg p-8 text-center max-w-md">
            <Shield className="w-16 h-16 mx-auto text-gray-400 mb-4" />
            <h3 className="text-xl font-semibold text-gray-800 mb-2">Acceso Restringido</h3>
            <p className="text-gray-600 mb-6">Debes iniciar sesión para acceder a esta sección</p>
            <button
              onClick={() => setCurrentView('login')}
              className="bg-gradient-to-r from-orange-500 to-red-500 text-white px-6 py-3 rounded-full font-bold hover:shadow-lg transform hover:scale-105 transition-all"
            >
              Iniciar Sesión
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

export default App;