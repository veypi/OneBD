# Form Creator Component

A modern, accessible form creation component built with Vue.js that provides a clean and professional interface for collecting user input.

## Features

- ✅ **Modern Design**: Clean, professional UI with smooth animations
- ✅ **Responsive**: Works perfectly on all device sizes
- ✅ **Accessible**: Full ARIA support and keyboard navigation
- ✅ **Form Validation**: Real-time validation with user-friendly error messages
- ✅ **Vue.js 3**: Built with Vue 3 Composition API
- ✅ **TypeScript Ready**: Fully typed for better development experience
- ✅ **Dark Mode**: Automatic dark mode support
- ✅ **No Dependencies**: Self-contained component with minimal footprint

## Fields

The component collects two fields:

1. **Name** (required): Text input for the item name
   - Minimum 2 characters
   - Maximum 100 characters
   - Required field validation

2. **Description** (optional): Textarea for additional details
   - Optional field
   - Expandable textarea
   - No character limit

## Usage

### Basic Usage

```vue
<template>
  <FormCreator 
    @submit="handleFormSubmit"
    @change="handleFormChange"
  />
</template>

<script setup>
import FormCreator from './components/FormCreator.vue'

const handleFormSubmit = (data) => {
  console.log('Form submitted:', data)
  // Handle form submission
  // Example: webapi.Post('/api/items', { json: data })
}

const handleFormChange = (data) => {
  console.log('Form changed:', data)
  // Handle form state changes
}
</script>
```

### With Reactive State

```vue
<template>
  <div>
    <FormCreator 
      @submit="handleFormSubmit"
      @change="handleFormChange"
    />
    
    <div v-if="formState.isValid">
      Form is ready to submit!
    </div>
  </div>
</template>

<script setup>
import { reactive } from 'vue'
import FormCreator from './components/FormCreator.vue'

const formState = reactive({
  name: '',
  description: '',
  isValid: false
})

const handleFormSubmit = (data) => {
  // Submit to API
  console.log('Submitting:', data)
}

const handleFormChange = (data) => {
  Object.assign(formState, data)
}
</script>
```

## Events

### @submit

Emitted when the form is successfully validated and submitted.

**Payload:**
```javascript
{
  name: string,      // Required field, trimmed
  description: string, // Optional field, trimmed
  timestamp: string   // ISO timestamp of submission
}
```

### @change

Emitted whenever form data changes, providing real-time form state.

**Payload:**
```javascript
{
  name: string,       // Current name value
  description: string, // Current description value
  isValid: boolean    // Whether form passes validation
}
```

## Validation Rules

### Name Field
- **Required**: Must not be empty
- **Minimum Length**: 2 characters
- **Maximum Length**: 100 characters
- **Validation Timing**: On blur and real-time error clearing

### Description Field
- **Optional**: No validation applied
- **Unlimited Length**: No character restrictions

## Styling

The component uses CSS custom properties (variables) for easy theming:

```css
.form-creator {
  --primary-color: #3b82f6;
  --primary-hover: #2563eb;
  --success-color: #10b981;
  --error-color: #ef4444;
  /* ... more variables */
}
```

### Customization

You can override the CSS variables to match your brand:

```css
.form-creator {
  --primary-color: #your-brand-color;
  --primary-hover: #your-brand-hover-color;
  --border-radius: 8px;
}
```

## Accessibility Features

- **ARIA Labels**: Proper labeling for screen readers
- **Focus Management**: Keyboard navigation support
- **Error Announcements**: Screen reader announcements for validation errors
- **Required Field Indicators**: Visual and semantic indicators
- **High Contrast Support**: Adapts to high contrast preferences
- **Reduced Motion**: Respects user motion preferences

## Browser Support

- ✅ Modern browsers (Chrome, Firefox, Safari, Edge)
- ✅ Mobile browsers (iOS Safari, Chrome Mobile)
- ✅ Responsive design for all screen sizes
- ✅ Automatic dark mode detection

## Demo

Open `frontend/demo.html` in your browser to see a live demonstration of the component.

## Integration with OneBD

This component can be easily integrated with the OneBD-generated API clients:

```javascript
import { webapi } from './path/to/generated/webapi'

const handleFormSubmit = async (data) => {
  try {
    const result = await webapi.Post('/api/items', { 
      json: data 
    })
    console.log('Item created:', result)
  } catch (error) {
    console.error('Failed to create item:', error)
  }
}
```

## File Structure

```
frontend/
├── components/
│   └── FormCreator.vue     # Main component
├── demo.html              # Live demo
└── README.md              # This documentation
```

## Development

The component is built with:
- Vue.js 3 Composition API
- Modern CSS with custom properties
- ES6+ JavaScript
- No external dependencies (except Vue.js)

### Component Structure

- **Template**: Semantic HTML5 with ARIA attributes
- **Script**: Vue 3 Composition API with reactive state management
- **Style**: Scoped CSS with modern design patterns

## License

This component is part of the OneBD project and follows the same license terms.