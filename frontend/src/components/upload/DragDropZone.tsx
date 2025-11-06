'use client';

import React, { useState, useCallback } from 'react';
import { useToast } from '@/contexts/ToastContext';

interface DragDropZoneProps {
  onFilesSelected: (files: File[]) => void;
  accept?: string;
  multiple?: boolean;
  maxFiles?: number;
  disabled?: boolean;
}

export function DragDropZone({
  onFilesSelected,
  accept = 'image/jpeg,image/jpg,image/png',
  multiple = false,
  maxFiles,
  disabled = false,
}: DragDropZoneProps) {
  const [isDragging, setIsDragging] = useState(false);
  const { warning } = useToast();
  
  const handleDrag = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
  }, []);
  
  const handleDragIn = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (e.dataTransfer.items && e.dataTransfer.items.length > 0 && !disabled) {
      setIsDragging(true);
    }
  }, [disabled]);
  
  const handleDragOut = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
  }, []);
  
  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    setIsDragging(false);
    
    if (disabled) return;
    
    const files = Array.from(e.dataTransfer.files);
    const acceptedTypes = accept.split(',').map(t => t.trim());
    
    const validFiles = files.filter(file => {
      return acceptedTypes.some(type => file.type.match(type.replace('*', '.*')));
    });
    
    const invalidFiles = files.length - validFiles.length;
    if (invalidFiles > 0) {
      warning(`Пропущено ${invalidFiles} файл(ов) с неподдерживаемым форматом`);
    }
    
    const filesToAdd = maxFiles ? validFiles.slice(0, maxFiles) : validFiles;
    
    if (maxFiles && validFiles.length > maxFiles) {
      warning(`Выбрано слишком много файлов. Добавлено только первые ${maxFiles}`);
    }
    
    if (filesToAdd.length > 0) {
      onFilesSelected(filesToAdd);
    } else if (files.length > 0) {
      warning('Нет подходящих файлов для загрузки');
    }
  }, [accept, disabled, maxFiles, onFilesSelected, warning]);
  
  const handleFileInput = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files.length > 0) {
      const files = Array.from(e.target.files);
      const filesToAdd = maxFiles ? files.slice(0, maxFiles) : files;
      
      if (maxFiles && files.length > maxFiles) {
        warning(`Выбрано слишком много файлов. Добавлено только первые ${maxFiles}`);
      }
      
      onFilesSelected(filesToAdd);
    }
  }, [maxFiles, onFilesSelected, warning]);
  
  return (
    <div
      className={`relative border-2 border-dashed rounded-xl p-8 text-center transition-all duration-200 ${
        isDragging
          ? 'border-indigo-500 bg-indigo-50 dark:border-indigo-400 dark:bg-indigo-900/30'
          : 'border-gray-300 dark:border-gray-600 hover:border-indigo-400 dark:hover:border-indigo-500 hover:bg-gray-50 dark:hover:bg-gray-700/30'
      } ${disabled ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}`}
      onDragEnter={handleDragIn}
      onDragLeave={handleDragOut}
      onDragOver={handleDrag}
      onDrop={handleDrop}
    >
      <input
        type="file"
        accept={accept}
        multiple={multiple}
        onChange={handleFileInput}
        disabled={disabled}
        className="absolute inset-0 w-full h-full opacity-0 cursor-pointer disabled:cursor-not-allowed"
      />
      
      <div className="space-y-4">
        <div className="flex justify-center">
          <svg
            className={`w-16 h-16 transition-colors ${
              isDragging ? 'text-indigo-500 dark:text-indigo-400' : 'text-gray-400 dark:text-gray-500'
            }`}
            fill="none"
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth="2"
            viewBox="0 0 24 24"
            stroke="currentColor"
          >
            <path d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
          </svg>
        </div>
        
        <div>
          <p className="text-lg font-medium text-gray-900 dark:text-gray-100">
            Перетащите файлы сюда или кликните для выбора
          </p>
          <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
            {multiple
              ? `Поддерживаются изображения JPG и PNG${maxFiles ? ` (до ${maxFiles} файлов)` : ''}`
              : 'Поддерживаются изображения JPG и PNG'}
          </p>
        </div>
      </div>
    </div>
  );
}

