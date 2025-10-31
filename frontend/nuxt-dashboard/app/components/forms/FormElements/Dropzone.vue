<script setup lang="ts">
import 'dropzone/dist/dropzone.css'

// Props
const props = defineProps<{
  uploadUrl?: string
}>()

const uploadUrl = props.uploadUrl ?? '/upload'

// Refs & state
const dropzoneForm = ref<HTMLFormElement | null>(null)
let dropzoneInstance: any = null

onMounted(async () => {
  await nextTick() // ensure form is in the DOM

  if (!dropzoneForm.value) return

  // Import Dropzone only on client
  const Dropzone = (await import('dropzone'!)).default
  // Prevent auto-instantiation on elements with .dropzone
  Dropzone.autoDiscover = false

  dropzoneInstance = new Dropzone(dropzoneForm.value, {
    url: uploadUrl,
    thumbnailWidth: 150,
    maxFilesize: 0.5, // MB
    acceptedFiles: 'image/jpeg,image/png,image/gif,image/webp,image/svg+xml',
    headers: { 'X-Custom-Header': 'example' },
    dictDefaultMessage: '', // we render our own message UI
    clickable: '.dz-message', // make the message area clickable
    init: function () {
      this.on('addedfile', (file: File) => {
        console.log('A file has been added', file)
      })
      this.on('success', (file: File, response: unknown) => {
        console.log('File successfully uploaded', file, response)
      })
      this.on('error', (file: File, error: unknown) => {
        console.error('Upload error', file, error)
      })
    }
  })
})

onBeforeUnmount(() => {
  // Destroy and free listeners
  dropzoneInstance?.destroy()
  dropzoneInstance = null
})
</script>

<template>
  <div class="file-uploader">
    <form
      ref="dropzoneForm"
      :action="uploadUrl"
      class="dropzone border border-dashed rounded-xl bg-gray-50 p-7 hover:border-brand-500
             dark:border-gray-700 dark:bg-gray-900 dark:hover:border-brand-500 lg:p-10"
      novalidate
      @submit.prevent
    >
      <div class="dz-message m-0!">
        <div class="mb-[22px] flex justify-center">
          <div
            class="flex h-[68px] w-[68px] items-center justify-center rounded-full bg-gray-200 text-gray-700 dark:bg-gray-800 dark:text-gray-400"
          >
            <svg class="fill-current" width="29" height="28" viewBox="0 0 29 28" xmlns="http://www.w3.org/2000/svg">
              <path
                fill-rule="evenodd" clip-rule="evenodd"
                d="M14.5019 3.91699c-.2167 0-.412 0.092-.5489.2389L8.5736 9.5319c-.293 .2928-.2931.7677 0 1.0606.2928.2929.7677.2931 1.0607 0l4.1179-4.1154V18.667c0 .4142.3358.75.75.75s.75-.3358.75-.75V6.4823l4.1144 4.1106c.2929.2928.7678.2926 1.0606-.0003.2928-.2929.2926-.7678-.0003-1.0606L15.0838 4.1938c-.1375-.1689-.3471-.2768-.5819-.2768ZM5.9163 18.667c0-.4142-.3358-.75-.75-.75s-.75.3358-.75.75v3.1667c0 1.2426 1.0074 2.25 2.25 2.25h15.6676c1.2427 0 2.25-1.0074 2.25-2.25V18.667c0-.4142-.3358-.75-.75-.75s-.75.3358-.75.75v3.1667c0 .4142-.3357.75-.75.75H6.6663c-.4142 0-.75-.3358-.75-.75V18.667Z"
              />
            </svg>
          </div>
        </div>

        <h4 class="mb-3 font-semibold text-gray-800 text-theme-xl dark:text-white/90">
          Drag & Drop File Here
        </h4>
        <span class="mx-auto mb-5 block w-full max-w-[290px] text-sm text-gray-700 dark:text-gray-400">
          Drag and drop your PNG, JPG, WebP, SVG images here or browse
        </span>

        <span class="font-medium underline cursor-pointer text-theme-sm text-brand-500">
          Browse File
        </span>
      </div>
    </form>
  </div>
</template>

<style>
.dropzone {
  border: 1px dashed #d0d5dd;
  transition: all 0.3s ease;
}
.dropzone:hover { border-color: #465fff; }
.dropzone .dz-preview { margin: 10px; }
.dropzone .dz-preview .dz-image { border-radius: 8px; }
.dropzone .dz-preview .dz-details { padding: 1em; }
.dropzone .dz-preview .dz-progress { height: 10px; }
.dropzone .dz-preview .dz-progress .dz-upload { background: #4f46e5; }

.dark .dropzone { background-color: #111827; border-color: #374151; }
.dark .dropzone:hover { border-color: #6366f1; }
</style>
