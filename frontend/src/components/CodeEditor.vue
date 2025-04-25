<template>
  <div ref="editorRef" class="code-editor" />
</template>

<script setup>
import { onMounted, onBeforeUnmount, ref, watch } from 'vue'
import { EditorState } from '@codemirror/state'
import { EditorView } from '@codemirror/view'
import { basicSetup } from '@codemirror/basic-setup'
import { python } from '@codemirror/lang-python'

const props = defineProps({
  modelValue: String,
  language: {
    type: String,
    default: 'python',
  },
})

const emit = defineEmits(['update:modelValue'])

const editorRef = ref()
let view = null

onMounted(() => {
  const lang = props.language === 'python' ? python() : []

  view = new EditorView({
    parent: editorRef.value,
    state: EditorState.create({
      doc: props.modelValue,
      extensions: [
        basicSetup,
        lang,
        EditorView.updateListener.of(update => {
          if (update.docChanged) {
            const value = update.state.doc.toString()
            emit('update:modelValue', value)
          }
        })
      ]
    })
  })
})

watch(() => props.modelValue, newVal => {
  if (view && newVal !== view.state.doc.toString()) {
    view.dispatch({
      changes: { from: 0, to: view.state.doc.length, insert: newVal }
    })
  }
})

onBeforeUnmount(() => {
  view?.destroy()
})
</script>

<style scoped>
.code-editor {
  border: 1px solid #ddd;
  border-radius: 4px;
  min-height: 300px;
  font-family: monospace;
  background-color: #f5f5f5;
}
</style>
