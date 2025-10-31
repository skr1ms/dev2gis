'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { useAuth } from '@/hooks/useAuth';
import { useToast } from '@/contexts/ToastContext';
import { Navbar } from '@/components/layout/Navbar';
import { Footer } from '@/components/layout/Footer';
import { DragDropZone } from '@/components/upload/DragDropZone';
import { ImagePreviewGrid } from '@/components/upload/ImagePreview';
import { ProgressBar } from '@/components/ui/ProgressBar';
import { Loading, Spinner } from '@/components/ui/Spinner';
import { validateImageFiles, validateBatchFiles } from '@/utils/validators';
import { api } from '@/utils/api';

export default function UploadPage() {
  const { user, loading: authLoading } = useAuth(true);
  const { success, error: showError, info } = useToast();
  const router = useRouter();
  
  const [files, setFiles] = useState<File[]>([]);
  const [uploading, setUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [mergeMethod, setMergeMethod] = useState<'medium' | 'max' | 'low'>('medium');
  const [fastMode, setFastMode] = useState(false);
  const [generateHeightmap, setGenerateHeightmap] = useState(true);
  const [generateOrthophoto, setGenerateOrthophoto] = useState(false);

  if (authLoading) {
    return <Loading text="Загрузка..." />;
  }

  if (!user) {
    return null;
  }

  const handleFilesSelected = (selectedFiles: File[]) => {
    const validationError = validateImageFiles(selectedFiles);
    if (validationError) {
      showError(validationError);
      return;
    }
    
    setFiles(prevFiles => [...prevFiles, ...selectedFiles]);
    success(`Добавлено ${selectedFiles.length} файл(ов)`);
  };

  const handleRemoveFile = (index: number) => {
    setFiles(prevFiles => prevFiles.filter((_, i) => i !== index));
    info('Файл удален');
  };

  const handleClearAll = () => {
    setFiles([]);
    info('Все файлы удалены');
  };

  const handleUploadSingle = async () => {
    if (files.length === 0) {
      showError('Выберите изображение для загрузки');
      return;
    }

    if (files.length > 1) {
      showError('Для одиночной обработки выберите только одно изображение');
      return;
    }

    setUploading(true);
    setUploadProgress(0);

    try {
      const formData = new FormData();
      formData.append('file', files[0]);

      setUploadProgress(30);

      const response = await api.post('/api/heightmaps/upload', formData);

      setUploadProgress(100);
      success('Изображение загружено! Обработка началась.');
      
      setTimeout(() => {
        router.push('/results');
      }, 1500);
    } catch (error: any) {
      showError(error.message || 'Ошибка загрузки');
      setUploadProgress(0);
    } finally {
      setUploading(false);
    }
  };

  const handleUploadBatch = async () => {
    const validationError = validateBatchFiles(files);
    if (validationError) {
      showError(validationError);
      return;
    }

    if (!generateHeightmap && !generateOrthophoto) {
      showError('Выберите хотя бы один тип генерации');
      return;
    }

    if (generateOrthophoto && files.length < 5) {
      showError('Для генерации ортофотоплана требуется минимум 5 изображений');
      return;
    }

    setUploading(true);
    setUploadProgress(0);

    try {
      const formData = new FormData();
      files.forEach(file => {
        formData.append('files', file);
      });
      formData.append('merge_method', mergeMethod);
      formData.append('fast_mode', String(fastMode));
      
      let generationMode = 'heightmap';
      if (generateHeightmap && generateOrthophoto) {
        generationMode = 'both';
      } else if (generateOrthophoto) {
        generationMode = 'orthophoto';
      }
      formData.append('generation_mode', generationMode);

      const progressInterval = setInterval(() => {
        setUploadProgress(prev => Math.min(prev + 5, 90));
      }, 200);

      const response = await api.post('/api/heightmaps/batch/upload', formData);

      clearInterval(progressInterval);
      setUploadProgress(100);
      success('Изображения загружены! Пакетная обработка началась.');
      
      setTimeout(() => {
        router.push('/results');
      }, 1500);
    } catch (error: any) {
      showError(error.message || 'Ошибка загрузки');
      setUploadProgress(0);
    } finally {
      setUploading(false);
    }
  };

  const canProcessSingle = files.length === 1;
  const canProcessBatch = files.length >= 5 && files.length <= 50;
  const processingType = files.length === 0 ? 'none' : files.length === 1 ? 'single' : files.length < 5 ? 'insufficient' : 'batch';

  return (
    <>
      <Navbar />
      <main className="flex-1 bg-gradient-surface">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="mb-8 animate-slideUp">
            <h1 className="text-4xl font-bold text-gray-900 dark:text-gray-100 mb-2">Загрузка изображений</h1>
            <p className="text-gray-600 dark:text-gray-300">
              Загрузите одно изображение для быстрой обработки или 5-20 изображений для точной фотограмметрии
            </p>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mb-8">
            <div className="bg-white dark:bg-gray-800 rounded-xl shadow-lg p-6 animate-slideUp stagger-delay-1 transition-colors">
              <div className="flex items-center mb-3">
                <div className="w-10 h-10 bg-indigo-100 dark:bg-indigo-900 rounded-full flex items-center justify-center mr-3">
                  <svg className="w-6 h-6 text-indigo-600 dark:text-indigo-400" fill="none" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" viewBox="0 0 24 24" stroke="currentColor">
                    <path d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                  </svg>
                </div>
                <h3 className="text-lg font-bold text-gray-900 dark:text-gray-100">Одиночная</h3>
              </div>
              <p className="text-sm text-gray-600 dark:text-gray-300 mb-3">1 изображение - быстрая оценка глубины (MiDaS)</p>
              <span className={`px-3 py-1 rounded-full text-xs font-medium ${
                processingType === 'single' ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200' : 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-300'
              }`}>
                {processingType === 'single' ? 'Выбрано' : '2-3 сек'}
              </span>
            </div>

            <div className="bg-white dark:bg-gray-800 rounded-xl shadow-lg p-6 animate-slideUp stagger-delay-2 transition-colors">
              <div className="flex items-center mb-3">
                <div className="w-10 h-10 bg-purple-100 dark:bg-purple-900 rounded-full flex items-center justify-center mr-3">
                  <svg className="w-6 h-6 text-purple-600 dark:text-purple-400" fill="none" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" viewBox="0 0 24 24" stroke="currentColor">
                    <path d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                  </svg>
                </div>
                <h3 className="text-lg font-bold text-gray-900 dark:text-gray-100">Пакетная</h3>
              </div>
              <p className="text-sm text-gray-600 dark:text-gray-300 mb-3">5-20 изображений - фотограмметрия (NodeODM)</p>
              <span className={`px-3 py-1 rounded-full text-xs font-medium ${
                processingType === 'batch' ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200' : 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-300'
              }`}>
                {processingType === 'batch' ? 'Выбрано' : '5-15 мин'}
              </span>
            </div>

            <div className="bg-white dark:bg-gray-800 rounded-xl shadow-lg p-6 animate-slideUp stagger-delay-3 transition-colors">
              <div className="flex items-center mb-3">
                <div className="w-10 h-10 bg-blue-100 dark:bg-blue-900 rounded-full flex items-center justify-center mr-3">
                  <svg className="w-6 h-6 text-blue-600 dark:text-blue-400" fill="none" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" viewBox="0 0 24 24" stroke="currentColor">
                    <path d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                </div>
                <h3 className="text-lg font-bold text-gray-900 dark:text-gray-100">Текущий выбор</h3>
              </div>
              <p className="text-sm text-gray-600 dark:text-gray-300 mb-3">Загружено: {files.length} файл(ов)</p>
              <span className={`px-3 py-1 rounded-full text-xs font-medium ${
                files.length === 0 ? 'bg-gray-100 text-gray-600 dark:bg-gray-700 dark:text-gray-300' :
                processingType === 'single' ? 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200' :
                processingType === 'batch' ? 'bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200' :
                'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200'
              }`}>
                {files.length === 0 ? 'Не загружено' :
                 processingType === 'single' ? 'Одиночная' :
                 processingType === 'batch' ? 'Пакетная' :
                 'Недостаточно (мин. 5)'}
              </span>
            </div>
          </div>

          <div className="bg-white dark:bg-gray-800 rounded-xl shadow-lg p-8 mb-8 animate-slideUp transition-colors">
            <DragDropZone
              onFilesSelected={handleFilesSelected}
              multiple={true}
              maxFiles={50}
              disabled={uploading}
            />

            {files.length > 0 && (
              <div className="mt-6">
                <div className="flex items-center justify-between mb-4">
                  <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100">
                    Выбранные файлы ({files.length})
                  </h3>
                  <button
                    onClick={handleClearAll}
                    disabled={uploading}
                    className="text-sm text-red-600 dark:text-red-400 hover:text-red-700 dark:hover:text-red-300 font-medium disabled:opacity-50 transition-colors"
                  >
                    Очистить все
                  </button>
                </div>

                <ImagePreviewGrid files={files} onRemove={handleRemoveFile} />
              </div>
            )}
          </div>

          {files.length >= 5 && (
            <div className="bg-white dark:bg-gray-800 rounded-xl shadow-lg p-6 mb-8 animate-slideUp transition-colors">
              <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">Тип генерации</h3>
              <p className="text-sm text-gray-600 dark:text-gray-300 mb-4">
                Выберите, что нужно сгенерировать на основе загруженных изображений
              </p>
              
              <div className="space-y-3 mb-6">
                <label className={`flex items-start cursor-pointer p-4 border-2 rounded-lg transition-all hover:border-indigo-300 dark:hover:border-indigo-600 ${generateHeightmap ? 'border-indigo-600 bg-indigo-50 dark:border-indigo-400 dark:bg-indigo-900/30' : 'border-gray-200 dark:border-gray-600'}`}>
                  <input
                    type="checkbox"
                    checked={generateHeightmap}
                    onChange={(e) => setGenerateHeightmap(e.target.checked)}
                    disabled={uploading}
                    className="w-5 h-5 text-indigo-600 bg-white dark:bg-gray-700 border-gray-300 dark:border-gray-600 rounded focus:ring-indigo-500 focus:ring-2 disabled:opacity-50 mt-0.5"
                  />
                  <div className="ml-3 flex-1">
                    <span className="text-sm font-medium text-gray-900 dark:text-gray-100 block">
                      Карта высот (DSM)
                    </span>
                    <p className="text-xs text-gray-600 dark:text-gray-400 mt-1">
                      Цифровая модель поверхности - визуализация рельефа и высот объектов
                    </p>
                  </div>
                </label>
                
                <label className={`flex items-start cursor-pointer p-4 border-2 rounded-lg transition-all hover:border-indigo-300 dark:hover:border-indigo-600 ${generateOrthophoto ? 'border-indigo-600 bg-indigo-50 dark:border-indigo-400 dark:bg-indigo-900/30' : 'border-gray-200 dark:border-gray-600'}`}>
                  <input
                    type="checkbox"
                    checked={generateOrthophoto}
                    onChange={(e) => setGenerateOrthophoto(e.target.checked)}
                    disabled={uploading}
                    className="w-5 h-5 text-indigo-600 bg-white dark:bg-gray-700 border-gray-300 dark:border-gray-600 rounded focus:ring-indigo-500 focus:ring-2 disabled:opacity-50 mt-0.5"
                  />
                  <div className="ml-3 flex-1">
                    <span className="text-sm font-medium text-gray-900 dark:text-gray-100 block">
                      Ортофотоплан
                    </span>
                    <p className="text-xs text-gray-600 dark:text-gray-400 mt-1">
                      Геометрически скорректированное изображение местности (требуется минимум 5 фото)
                    </p>
                  </div>
                </label>
              </div>
              
              {generateOrthophoto && files.length < 5 && (
                <div className="bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg p-4 mb-6">
                  <div className="flex">
                    <svg className="w-5 h-5 text-yellow-600 dark:text-yellow-400 flex-shrink-0" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                    </svg>
                    <div className="ml-3">
                      <p className="text-sm text-yellow-800 dark:text-yellow-200 font-medium">
                        Недостаточно изображений
                      </p>
                      <p className="text-xs text-yellow-700 dark:text-yellow-300 mt-1">
                        Для генерации ортофотоплана загрузите минимум 5 фотографий. Сейчас загружено: {files.length}
                      </p>
                    </div>
                  </div>
                </div>
              )}
            </div>
          )}

          {files.length >= 5 && (
            <div className="bg-white dark:bg-gray-800 rounded-xl shadow-lg p-6 mb-8 animate-slideUp transition-colors">
              <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">Настройки пакетной обработки</h3>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">
                  Качество обработки
                </label>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                  <button
                    onClick={() => setMergeMethod('low')}
                    disabled={uploading}
                    className={`p-4 border-2 rounded-lg transition-all ${
                      mergeMethod === 'low'
                        ? 'border-indigo-600 bg-indigo-50 dark:border-indigo-400 dark:bg-indigo-900/30'
                        : 'border-gray-200 dark:border-gray-600 hover:border-gray-300 dark:hover:border-gray-500'
                    } disabled:opacity-50`}
                  >
                    <div className="font-semibold text-gray-900 dark:text-gray-100">Низкое (Low)</div>
                    <div className="text-sm text-gray-600 dark:text-gray-300 mt-1">Быстрая обработка</div>
                  </button>
                  
                  <button
                    onClick={() => setMergeMethod('medium')}
                    disabled={uploading}
                    className={`p-4 border-2 rounded-lg transition-all ${
                      mergeMethod === 'medium'
                        ? 'border-indigo-600 bg-indigo-50 dark:border-indigo-400 dark:bg-indigo-900/30'
                        : 'border-gray-200 dark:border-gray-600 hover:border-gray-300 dark:hover:border-gray-500'
                    } disabled:opacity-50`}
                  >
                    <div className="font-semibold text-gray-900 dark:text-gray-100">Среднее (Medium)</div>
                    <div className="text-sm text-gray-600 dark:text-gray-300 mt-1">Оптимальный баланс</div>
                  </button>
                  
                  <button
                    onClick={() => setMergeMethod('max')}
                    disabled={uploading}
                    className={`p-4 border-2 rounded-lg transition-all ${
                      mergeMethod === 'max'
                        ? 'border-indigo-600 bg-indigo-50 dark:border-indigo-400 dark:bg-indigo-900/30'
                        : 'border-gray-200 dark:border-gray-600 hover:border-gray-300 dark:hover:border-gray-500'
                    } disabled:opacity-50`}
                  >
                    <div className="font-semibold text-gray-900 dark:text-gray-100">Ультра (Ultra)</div>
                    <div className="text-sm text-gray-600 dark:text-gray-300 mt-1">Максимальное качество</div>
                  </button>
                </div>
              </div>
              
              <div className="mt-6 pt-6 border-t border-gray-200 dark:border-gray-700">
                <label className="flex items-center cursor-pointer">
                  <input
                    type="checkbox"
                    checked={fastMode}
                    onChange={(e) => setFastMode(e.target.checked)}
                    disabled={uploading}
                    className="w-5 h-5 text-indigo-600 bg-white dark:bg-gray-700 border-gray-300 dark:border-gray-600 rounded focus:ring-indigo-500 focus:ring-2 disabled:opacity-50"
                  />
                  <div className="ml-3">
                    <span className="text-sm font-medium text-gray-900 dark:text-gray-100">
                      Быстрая генерация
                    </span>
                    <p className="text-xs text-gray-600 dark:text-gray-400 mt-1">
                      Пропускает 3D модель и детальную обработку. Ускоряет генерацию в 2-3 раза.
                    </p>
                  </div>
                </label>
              </div>
            </div>
          )}

          {uploading && (
            <div className="bg-white dark:bg-gray-800 rounded-xl shadow-lg p-8 mb-8 animate-slideUp transition-colors">
              <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-4">Загрузка файлов...</h3>
              <ProgressBar progress={uploadProgress} showLabel size="lg" />
              <p className="text-sm text-gray-600 dark:text-gray-300 mt-4 text-center">
                Пожалуйста, дождитесь завершения загрузки
              </p>
            </div>
          )}

          <div className="flex gap-4">
            <button
              onClick={handleUploadSingle}
              disabled={!canProcessSingle || uploading}
              className="flex-1 px-6 py-4 bg-indigo-600 text-white rounded-lg font-semibold hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-300 hover:shadow-xl flex items-center justify-center"
            >
              {uploading ? (
                <Spinner size="sm" className="mr-2" />
              ) : (
                <svg className="w-5 h-5 mr-2" fill="none" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" viewBox="0 0 24 24" stroke="currentColor">
                  <path d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                </svg>
              )}
              Обработать одно изображение
            </button>

            <button
              onClick={handleUploadBatch}
              disabled={!canProcessBatch || uploading}
              className="flex-1 px-6 py-4 bg-purple-600 text-white rounded-lg font-semibold hover:bg-purple-700 disabled:opacity-50 disabled:cursor-not-allowed transition-all duration-300 hover:shadow-xl flex items-center justify-center"
            >
              {uploading ? (
                <Spinner size="sm" className="mr-2" />
              ) : (
                <svg className="w-5 h-5 mr-2" fill="none" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" viewBox="0 0 24 24" stroke="currentColor">
                  <path d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
                </svg>
              )}
              Обработать пакет ({files.length})
            </button>
          </div>
        </div>
      </main>
      <Footer />
    </>
  );
}
