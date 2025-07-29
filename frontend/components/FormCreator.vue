<template>
  <div class="form-creator">
    <div class="form-container">
      <h2 class="form-title">Create New Item</h2>
      <form @submit.prevent="handleSubmit" class="form" novalidate>
        <!-- Name Field -->
        <div class="form-group">
          <label for="name" class="form-label required">
            Name <span class="required-indicator">*</span>
          </label>
          <input
            id="name"
            v-model="formData.name"
            type="text"
            class="form-input"
            :class="{ 
              'error': errors.name, 
              'success': !errors.name && formData.name.trim() 
            }"
            placeholder="Enter name"
            @blur="validateName"
            @input="clearError('name')"
            aria-describedby="name-error"
            aria-required="true"
          />
          <transition name="error-fade">
            <div v-if="errors.name" id="name-error" class="error-message" role="alert">
              {{ errors.name }}
            </div>
          </transition>
        </div>

        <!-- Description Field -->
        <div class="form-group">
          <label for="description" class="form-label">
            Description
          </label>
          <textarea
            id="description"
            v-model="formData.description"
            class="form-textarea"
            :class="{ 'success': formData.description.trim() }"
            placeholder="Enter description (optional)"
            rows="4"
            @input="clearError('description')"
            aria-describedby="description-help"
          ></textarea>
          <div id="description-help" class="field-help">
            Optional field to provide additional details
          </div>
        </div>

        <!-- Submit Button -->
        <div class="form-actions">
          <button
            type="submit"
            class="submit-button"
            :class="{ 
              'loading': isSubmitting,
              'disabled': !isFormValid || isSubmitting
            }"
            :disabled="!isFormValid || isSubmitting"
            aria-describedby="submit-help"
          >
            <span v-if="!isSubmitting" class="button-text">
              <svg class="button-icon" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M12 5v14M5 12h14"/>
              </svg>
              Create Item
            </span>
            <span v-else class="button-text">
              <svg class="loading-spinner" width="16" height="16" viewBox="0 0 24 24">
                <circle cx="12" cy="12" r="3" fill="currentColor"/>
              </svg>
              Creating...
            </span>
          </button>
          <div id="submit-help" class="submit-help">
            All required fields must be completed
          </div>
        </div>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, reactive } from 'vue'

// Define emits
const emit = defineEmits(['submit', 'change'])

// Reactive form data
const formData = reactive({
  name: '',
  description: ''
})

// Form state
const errors = reactive({
  name: '',
  description: ''
})
const isSubmitting = ref(false)

// Computed properties
const isFormValid = computed(() => {
  return formData.name.trim().length > 0 && !errors.name
})

// Validation functions
const validateName = () => {
  const name = formData.name.trim()
  if (!name) {
    errors.name = 'Name is required'
    return false
  }
  if (name.length < 2) {
    errors.name = 'Name must be at least 2 characters long'
    return false
  }
  if (name.length > 100) {
    errors.name = 'Name must be less than 100 characters'
    return false
  }
  errors.name = ''
  return true
}

const clearError = (field) => {
  if (errors[field]) {
    errors[field] = ''
  }
}

// Form submission
const handleSubmit = async () => {
  // Validate all fields
  const isNameValid = validateName()
  
  if (!isNameValid) {
    // Focus on first error field
    const firstErrorField = document.querySelector('.form-input.error, .form-textarea.error')
    if (firstErrorField) {
      firstErrorField.focus()
    }
    return
  }

  isSubmitting.value = true

  try {
    // Simulate API call delay
    await new Promise(resolve => setTimeout(resolve, 1000))
    
    // Emit the form data
    const submitData = {
      name: formData.name.trim(),
      description: formData.description.trim()
    }
    
    emit('submit', submitData)
    
    // Reset form after successful submission
    formData.name = ''
    formData.description = ''
    errors.name = ''
    errors.description = ''
    
  } catch (error) {
    console.error('Form submission error:', error)
  } finally {
    isSubmitting.value = false
  }
}

// Emit changes for parent component reactivity
const emitChange = () => {
  emit('change', {
    name: formData.name,
    description: formData.description,
    isValid: isFormValid.value
  })
}

// Watch for changes and emit
import { watch } from 'vue'
watch(formData, emitChange, { deep: true })
</script>

<style scoped>
/* CSS Custom Properties for theming */
.form-creator {
  --primary-color: #3b82f6;
  --primary-hover: #2563eb;
  --primary-active: #1d4ed8;
  --success-color: #10b981;
  --error-color: #ef4444;
  --warning-color: #f59e0b;
  --neutral-50: #f9fafb;
  --neutral-100: #f3f4f6;
  --neutral-200: #e5e7eb;
  --neutral-300: #d1d5db;
  --neutral-400: #9ca3af;
  --neutral-500: #6b7280;
  --neutral-600: #4b5563;
  --neutral-700: #374151;
  --neutral-800: #1f2937;
  --neutral-900: #111827;
  --shadow-sm: 0 1px 2px 0 rgb(0 0 0 / 0.05);
  --shadow-md: 0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1);
  --shadow-lg: 0 10px 15px -3px rgb(0 0 0 / 0.1), 0 4px 6px -4px rgb(0 0 0 / 0.1);
  --border-radius: 0.5rem;
  --border-radius-sm: 0.375rem;
  --transition-base: all 0.2s ease-in-out;
  --transition-colors: color 0.15s ease-in-out, background-color 0.15s ease-in-out, border-color 0.15s ease-in-out;
}

/* Container and Layout */
.form-creator {
  max-width: 600px;
  margin: 0 auto;
  padding: 1.5rem;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  line-height: 1.5;
  color: var(--neutral-800);
}

.form-container {
  background: white;
  border-radius: var(--border-radius);
  box-shadow: var(--shadow-lg);
  overflow: hidden;
  border: 1px solid var(--neutral-200);
}

.form-title {
  margin: 0;
  padding: 2rem 2rem 1rem;
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--neutral-900);
  text-align: center;
  background: linear-gradient(135deg, var(--neutral-50) 0%, white 100%);
  border-bottom: 1px solid var(--neutral-100);
}

.form {
  padding: 2rem;
}

/* Form Groups */
.form-group {
  margin-bottom: 1.5rem;
  position: relative;
}

.form-group:last-of-type {
  margin-bottom: 2rem;
}

/* Labels */
.form-label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
  color: var(--neutral-700);
  font-size: 0.875rem;
  letter-spacing: 0.025em;
}

.form-label.required {
  position: relative;
}

.required-indicator {
  color: var(--error-color);
  font-weight: 600;
  margin-left: 0.125rem;
}

/* Input Styles */
.form-input,
.form-textarea {
  width: 100%;
  padding: 0.75rem 1rem;
  border: 2px solid var(--neutral-200);
  border-radius: var(--border-radius-sm);
  font-size: 1rem;
  font-family: inherit;
  background-color: white;
  transition: var(--transition-colors);
  box-sizing: border-box;
}

.form-input:focus,
.form-textarea:focus {
  outline: none;
  border-color: var(--primary-color);
  box-shadow: 0 0 0 3px rgb(59 130 246 / 0.1);
}

.form-input:hover:not(:focus):not(.error),
.form-textarea:hover:not(:focus):not(.error) {
  border-color: var(--neutral-300);
}

.form-input.success,
.form-textarea.success {
  border-color: var(--success-color);
  background-color: rgb(240 253 244);
}

.form-input.error,
.form-textarea.error {
  border-color: var(--error-color);
  background-color: rgb(254 242 242);
}

.form-textarea {
  resize: vertical;
  min-height: 100px;
  line-height: 1.5;
}

/* Placeholder styles */
.form-input::placeholder,
.form-textarea::placeholder {
  color: var(--neutral-400);
  opacity: 1;
}

/* Field Help Text */
.field-help {
  margin-top: 0.375rem;
  font-size: 0.875rem;
  color: var(--neutral-500);
  line-height: 1.4;
}

/* Error Messages */
.error-message {
  margin-top: 0.375rem;
  font-size: 0.875rem;
  color: var(--error-color);
  font-weight: 500;
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.error-message::before {
  content: '⚠';
  font-size: 1rem;
}

/* Error transition */
.error-fade-enter-active,
.error-fade-leave-active {
  transition: all 0.3s ease;
}

.error-fade-enter-from,
.error-fade-leave-to {
  opacity: 0;
  transform: translateY(-0.5rem);
}

/* Form Actions */
.form-actions {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  align-items: stretch;
}

.submit-button {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.875rem 1.5rem;
  background: linear-gradient(135deg, var(--primary-color) 0%, var(--primary-hover) 100%);
  color: white;
  border: none;
  border-radius: var(--border-radius-sm);
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: var(--transition-base);
  box-shadow: var(--shadow-sm);
  position: relative;
  overflow: hidden;
}

.submit-button:hover:not(.disabled) {
  background: linear-gradient(135deg, var(--primary-hover) 0%, var(--primary-active) 100%);
  box-shadow: var(--shadow-md);
  transform: translateY(-1px);
}

.submit-button:active:not(.disabled) {
  transform: translateY(0);
  box-shadow: var(--shadow-sm);
}

.submit-button.disabled {
  background: var(--neutral-300);
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

.submit-button.loading {
  cursor: wait;
}

.button-text {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.button-icon {
  transition: transform 0.2s ease;
}

.submit-button:hover:not(.disabled) .button-icon {
  transform: scale(1.1);
}

.loading-spinner {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.submit-help {
  text-align: center;
  font-size: 0.75rem;
  color: var(--neutral-500);
  font-style: italic;
}

/* Responsive Design */
@media (max-width: 640px) {
  .form-creator {
    padding: 1rem;
    max-width: 100%;
  }
  
  .form-container {
    border-radius: var(--border-radius-sm);
  }
  
  .form-title {
    padding: 1.5rem 1.5rem 1rem;
    font-size: 1.25rem;
  }
  
  .form {
    padding: 1.5rem;
  }
  
  .form-input,
  .form-textarea {
    padding: 0.625rem 0.875rem;
    font-size: 0.875rem;
  }
  
  .submit-button {
    padding: 0.75rem 1.25rem;
    font-size: 0.875rem;
  }
}

/* Accessibility improvements */
@media (prefers-reduced-motion: reduce) {
  * {
    animation-duration: 0.01ms !important;
    animation-iteration-count: 1 !important;
    transition-duration: 0.01ms !important;
  }
}

/* Focus indicators for keyboard navigation */
.form-input:focus-visible,
.form-textarea:focus-visible,
.submit-button:focus-visible {
  outline: 2px solid var(--primary-color);
  outline-offset: 2px;
}

/* High contrast mode support */
@media (prefers-contrast: high) {
  .form-container {
    border: 2px solid var(--neutral-800);
  }
  
  .form-input,
  .form-textarea {
    border-width: 2px;
  }
}

/* Dark mode support */
@media (prefers-color-scheme: dark) {
  .form-creator {
    --neutral-50: #1f2937;
    --neutral-100: #374151;
    --neutral-200: #4b5563;
    --neutral-300: #6b7280;
    --neutral-400: #9ca3af;
    --neutral-500: #d1d5db;
    --neutral-600: #e5e7eb;
    --neutral-700: #f3f4f6;
    --neutral-800: #f9fafb;
    --neutral-900: #ffffff;
  }
  
  .form-container {
    background: var(--neutral-100);
    border-color: var(--neutral-200);
  }
  
  .form-input,
  .form-textarea {
    background-color: var(--neutral-50);
    color: var(--neutral-800);
  }
  
  .form-input.success,
  .form-textarea.success {
    background-color: rgb(6 78 59);
  }
  
  .form-input.error,
  .form-textarea.error {
    background-color: rgb(127 29 29);
  }
}
</style>