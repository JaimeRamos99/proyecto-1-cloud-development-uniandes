const BASE_URL = 'http://localhost:80'; 

class ApiService {
    constructor() {
        this.baseURL = BASE_URL;
        this.token = localStorage.getItem('access_token');
    }

    async request(endpoint, options = {}) {
        const url = `${this.baseURL}${endpoint}`;
        const config = {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers,
            },
            ...options,
        };

        const token = this.token || localStorage.getItem('access_token');
        if (token) {
            config.headers.Authorization = `Bearer ${token}`;
        }
        console.log('🔑 Using token:', token);  
        console.log('📡 Headers:', config.headers);

        try {
            const response = await fetch(url, config);
            console.log('📡 Response Status:', response.status);
            
            if (response.status === 204) {
                return {};
            }
            
            // Check if response is JSON
            const contentType = response.headers.get('content-type');
            if (!contentType || !contentType.includes('application/json')) {
                const text = await response.text();
                console.error('❌ Non-JSON Response:', text);
                throw new Error(`Server returned non-JSON response: ${text.substring(0, 100)}`);
            }
            
            const data = await response.json();
            console.log('✅ Response Data:', data);

            if (!response.ok) {
                throw new Error(data.error || `HTTP error! status: ${response.status}`);
            }

            return data;
        } catch (error) {
            console.error('❌ API request failed:', error);
            console.error('❌ Failed URL:', url);
            throw error;
        }
    }

    async signup(userData) {
        const response = await this.request('/api/auth/signup', {
            method: 'POST',
            body: JSON.stringify({
                first_name: userData.firstName,
                last_name: userData.lastName,
                email: userData.email,
                password1: userData.password,
                password2: userData.confirmPassword,
                city: userData.city,
                country: userData.country || 'Colombia',
            }),
        });
        return response;
    }

    async login(email, password) {
        const response = await this.request('/api/auth/login', {
            method: 'POST',
            body: JSON.stringify({
                email,
                password,
            }),
        });

        if (response.token) {
            this.token = response.token;
            localStorage.setItem('access_token', response.token);
        }

        return response;
    }

    async getProfile() {
        return await this.request('/api/auth/profile');
    }

    async uploadVideo(title, file, isPublic = true) {
        console.log('📤 Uploading video:', { title, fileName: file.name, isPublic });
        
        const formData = new FormData();
        formData.append('title', title);
        formData.append('video_file', file); // ✅ Matches backend expectation
        formData.append('is_public', isPublic.toString());

        // Don't use this.request for FormData - handle fetch directly
        const response = await fetch(`${this.baseURL}/api/videos/upload`, {
            method: 'POST',
            headers: {
                Authorization: `Bearer ${this.token}`,
                // Don't set Content-Type for FormData - browser sets it with boundary
            },
            body: formData,
        });

        console.log('📤 Upload Response Status:', response.status);

        if (!response.ok) {
            const errorText = await response.text();
            console.error('📤 Upload Error:', errorText);
            
            try {
                const errorJson = JSON.parse(errorText);
                throw new Error(errorJson.error || 'Upload failed');
            } catch (parseError) {
                throw new Error(errorText || 'Upload failed');
            }
        }

        return await response.json();
    }

    async getMyVideos() {
        console.log('📹 Fetching my videos...');
        const result = await this.request('/api/videos/'); 
        return Array.isArray(result) ? result : [];
    }

    async getPublicVideos() {
        const result = await this.request('/api/public/videos/'); 
        console.log(' Public Videos Result:', result);
        return Array.isArray(result) ? result : [];
    }

    async getVideo(videoId) {
        return await this.request(`/api/videos/${videoId}`);
    }

    async deleteVideo(videoId) {
        return await this.request(`/api/videos/${videoId}`, {
            method: 'DELETE',
        });
    }

    async voteVideo(videoId) {
        console.log('👍 Voting for video:', videoId);
        return await this.request(`/api/public/videos/${videoId}/vote`, {
            method: 'POST',
        });
    }

    async unvoteVideo(videoId) {
        return await this.request(`/api/public/videos/${videoId}/vote`, {
            method: 'DELETE',
        });
    }

    async getTopRankings(limit = 10, city = '') {
        console.log('🏆 Fetching rankings:', { limit, city });
        const params = new URLSearchParams();
        if (limit) params.append('page_size', limit);
        if (city && city !== 'todas') params.append('city', city);
        params.append('page', '1');
        
        const query = params.toString() ? `?${params.toString()}` : '';
        const result = await this.request(`/api/public/rankings${query}`);
        return result || { rankings: [], total: 0, page: 1, page_size: limit };
    }

    async getPlayerRanking(userId) {
        return await this.request(`/api/public/rankings/${userId}`);
    }

    async refreshRankings() {
        return await this.request('/api/public/rankings/refresh', {
            method: 'POST',
        });
    }

    async logout() {
        if (this.token) {
            try {
                await this.request('/api/auth/logout', {
                    method: 'POST',
                });
            } catch (error) {
                console.error('Server logout failed:', error);
            }
        }
        
        this.token = null;
        localStorage.removeItem('access_token');
    }

    isAuthenticated() {
        return !!this.token;
    }

    async healthCheck() {
        return await this.request('/api/health');
    }

    getVideoStreamUrl(videoId) {
        // Your Go backend should have an endpoint like this for streaming
        return `${this.baseURL}/api/public/videos/${videoId}/stream`;
    }

    getVideoUrl(video) {
        // If your backend returns direct URLs in the video object
        return this.getVideoStreamUrl(video.video_id);
    }

    async findServerPort() {
        const ports = [8080, 8000, 3000, 8081, 9000];
        console.log('🔍 Looking for Go server on ports:', ports);
        
        for (const port of ports) {
            try {
                const testUrl = `http://localhost:${port}/api/health`;
                console.log('🧪 Testing:', testUrl);
                
                const response = await fetch(testUrl);
                if (response.ok) {
                    console.log('✅ Found server on port:', port);
                    return port;
                }
            } catch (error) {
                console.log('❌ Port', port, 'failed:', error.message);
            }
        }
        
        console.log('❌ No server found on any common port');
        return null;
    }

    async testConnection() {
        try {
            console.log('🧪 Testing connection to:', this.baseURL);
            const health = await this.healthCheck();
            console.log('✅ Health check passed:', health);
            return true;
        } catch (error) {
            console.error('❌ Connection test failed:', error);
            
            // Try to find the server on other ports
            const foundPort = await this.findServerPort();
            if (foundPort) {
                console.log(`💡 Suggestion: Update BASE_URL to http://localhost:${foundPort}`);
            }
            
            return false;
        }
    }

    async getTopRankings(limit = 10, city = '') {
        console.log('Fetching rankings with params:', { limit, city });
        
        const params = new URLSearchParams();
        if (limit) params.append('page_size', limit);
        if (city && city !== 'todas') params.append('city', city);
        params.append('page', '1');
        
        const query = params.toString() ? `?${params.toString()}` : '';
        const endpoint = `/api/public/rankings${query}`;
        
        try {
            console.log('Fetching from:', `${this.baseURL}${endpoint}`);
            const result = await this.request(endpoint);
            console.log('Rankings API response:', result);
            
            // Handle the response structure from your Go backend
            if (result && typeof result === 'object') {
                const rankings = result.rankings || [];
                const pagination = result.pagination || {};
                
                console.log('Raw rankings from API:', rankings);
                console.log('Number of rankings:', rankings.length);
                
                return {
                    rankings: Array.isArray(rankings) ? rankings : [],
                    total: pagination.total_items || 0,
                    page: pagination.current_page || 1,
                    page_size: pagination.page_size || limit,
                    total_pages: pagination.total_pages || 0
                };
            }
            
            console.log('Unexpected response structure, returning empty');
            return { rankings: [], total: 0, page: 1, page_size: limit };
            
        } catch (error) {
            console.error('Rankings request failed:', error);
            return { rankings: [], total: 0, page: 1, page_size: limit };
        }
    }

    // New refreshRankings method
    async refreshRankings() {
        try {
            console.log('Refreshing player rankings...');
            const result = await this.request('/api/public/rankings/refresh', 'POST');
            console.log('Rankings refresh result:', result);
            return result;
        } catch (error) {
            console.error('Failed to refresh rankings:', error);
            throw error;
        }
    }

    // Optional: Method to get a specific player's ranking
    async getPlayerRanking(userId) {
        try {
            console.log('Fetching ranking for user:', userId);
            const result = await this.request(`/api/public/rankings/player/${userId}`);
            console.log('Player ranking result:', result);
            return result;
        } catch (error) {
            console.error('Failed to get player ranking:', error);
            throw error;
        }
    }
}

export default new ApiService();