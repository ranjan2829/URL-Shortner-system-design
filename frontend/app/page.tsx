'use client';

import { useState } from 'react';

interface ShortenResponse {
  short_url: string;
  short_code: string;
  original_url: string;
  expires_at?: string;
}

export default function Home() {
  const [url, setUrl] = useState('');
  const [expiresIn, setExpiresIn] = useState<number | ''>('');
  const [result, setResult] = useState<ShortenResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [copied, setCopied] = useState(false);

  const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    setResult(null);
    setCopied(false);

    try {
      const body: any = { url };
      if (expiresIn) {
        body.expires_in = expiresIn;
      }

      const response = await fetch(`${API_BASE}/api/v1/shorten`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(body),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Failed to shorten URL');
      }

      const data: ShortenResponse = await response.json();
      setResult(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Something went wrong');
    } finally {
      setLoading(false);
    }
  };

  const copyToClipboard = () => {
    if (result?.short_url) {
      navigator.clipboard.writeText(result.short_url);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 via-white to-purple-50 dark:from-gray-900 dark:via-gray-800 dark:to-gray-900">
      <div className="container mx-auto px-4 py-16">
        <div className="max-w-2xl mx-auto">
          {/* Header */}
          <div className="text-center mb-12">
            <h1 className="text-5xl font-bold text-gray-900 dark:text-white mb-4">
              URL Shortener
            </h1>
            <p className="text-lg text-gray-600 dark:text-gray-300">
              Shorten your long URLs instantly. Free, fast, and reliable.
            </p>
          </div>

          {/* Form */}
          <div className="bg-white dark:bg-gray-800 rounded-2xl shadow-xl p-8 mb-8">
            <form onSubmit={handleSubmit} className="space-y-6">
              <div>
                <label
                  htmlFor="url"
                  className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                >
                  Enter URL to shorten
                </label>
                <input
                  type="url"
                  id="url"
                  value={url}
                  onChange={(e) => setUrl(e.target.value)}
                  placeholder="https://example.com/very/long/url"
                  required
                  className="w-full px-4 py-3 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent dark:bg-gray-700 dark:text-white outline-none transition"
                />
              </div>

              <div>
                <label
                  htmlFor="expires"
                  className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2"
                >
                  Expires in (hours) - Optional
                </label>
                <input
                  type="number"
                  id="expires"
                  value={expiresIn}
                  onChange={(e) => setExpiresIn(e.target.value ? Number(e.target.value) : '')}
                  placeholder="24"
                  min="1"
                  className="w-full px-4 py-3 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent dark:bg-gray-700 dark:text-white outline-none transition"
                />
              </div>

              <button
                type="submit"
                disabled={loading || !url}
                className="w-full bg-blue-600 hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed text-white font-semibold py-3 px-6 rounded-lg transition duration-200 shadow-lg hover:shadow-xl"
              >
                {loading ? 'Shortening...' : 'Shorten URL'}
              </button>
            </form>
          </div>

          {/* Error Message */}
          {error && (
            <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4 mb-6">
              <p className="text-red-800 dark:text-red-200">{error}</p>
            </div>
          )}

          {/* Result */}
          {result && (
            <div className="bg-white dark:bg-gray-800 rounded-2xl shadow-xl p-8">
              <h2 className="text-2xl font-bold text-gray-900 dark:text-white mb-6">
                Your Short URL is Ready! 🎉
              </h2>
              
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                    Short URL
                  </label>
                  <div className="flex gap-2">
                    <input
                      type="text"
                      value={result.short_url}
                      readOnly
                      className="flex-1 px-4 py-3 border border-gray-300 dark:border-gray-600 rounded-lg bg-gray-50 dark:bg-gray-700 text-gray-900 dark:text-white font-mono"
                    />
                    <button
                      onClick={copyToClipboard}
                      className="px-6 py-3 bg-blue-600 hover:bg-blue-700 text-white font-semibold rounded-lg transition duration-200"
                    >
                      {copied ? '✓ Copied!' : 'Copy'}
                    </button>
                  </div>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4 pt-4 border-t border-gray-200 dark:border-gray-700">
                  <div>
                    <p className="text-sm text-gray-600 dark:text-gray-400 mb-1">Original URL</p>
                    <p className="text-sm font-mono text-gray-900 dark:text-white break-all">
                      {result.original_url}
                    </p>
                  </div>
                  <div>
                    <p className="text-sm text-gray-600 dark:text-gray-400 mb-1">Short Code</p>
                    <p className="text-sm font-mono text-gray-900 dark:text-white">
                      {result.short_code}
                    </p>
                  </div>
                  {result.expires_at && (
                    <div>
                      <p className="text-sm text-gray-600 dark:text-gray-400 mb-1">Expires At</p>
                      <p className="text-sm text-gray-900 dark:text-white">
                        {new Date(result.expires_at).toLocaleString()}
          </p>
        </div>
                  )}
                </div>

                <div className="pt-4">
          <a
                    href={result.short_url}
            target="_blank"
            rel="noopener noreferrer"
                    className="inline-block px-6 py-3 bg-green-600 hover:bg-green-700 text-white font-semibold rounded-lg transition duration-200"
                  >
                    Test Short URL →
                  </a>
                </div>
              </div>
            </div>
          )}

          {/* Features */}
          <div className="mt-12 grid grid-cols-1 md:grid-cols-3 gap-6">
            <div className="text-center p-6 bg-white dark:bg-gray-800 rounded-xl shadow-lg">
              <div className="text-4xl mb-4">⚡</div>
              <h3 className="font-semibold text-gray-900 dark:text-white mb-2">Fast</h3>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                Instant URL shortening
              </p>
            </div>
            <div className="text-center p-6 bg-white dark:bg-gray-800 rounded-xl shadow-lg">
              <div className="text-4xl mb-4">🔒</div>
              <h3 className="font-semibold text-gray-900 dark:text-white mb-2">Secure</h3>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                Safe and reliable links
              </p>
            </div>
            <div className="text-center p-6 bg-white dark:bg-gray-800 rounded-xl shadow-lg">
              <div className="text-4xl mb-4">📊</div>
              <h3 className="font-semibold text-gray-900 dark:text-white mb-2">Analytics</h3>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                Track click statistics
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
