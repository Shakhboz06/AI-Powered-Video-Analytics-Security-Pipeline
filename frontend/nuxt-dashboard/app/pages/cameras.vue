<template>
  <div class="space-y-6">
    <header class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
      <div>
        <h1 class="text-xl font-semibold text-gray-100">Cameras</h1>
        <p class="mt-1 text-sm text-gray-500">Manage configured camera sources used by monitoring, zones, alerts, and analytics.</p>
      </div>
      <button
        type="button"
        class="rounded-lg bg-teal-600 px-4 py-2 text-sm font-medium text-white hover:bg-teal-500"
        @click="openCreate"
      >
        Add camera
      </button>
    </header>

    <div v-if="pageError" class="rounded-lg border border-red-900/50 bg-red-950/30 px-4 py-3 text-sm text-red-300">
      {{ pageError }}
    </div>

    <section class="rounded-xl border border-gray-800 bg-[#12181f] overflow-hidden">
      <div class="border-b border-gray-800 px-4 py-3 flex items-center justify-between">
        <h2 class="text-sm font-medium text-gray-400 uppercase tracking-wide">Camera list</h2>
        <button
          type="button"
          class="text-xs text-teal-300 hover:underline disabled:text-gray-600"
          :disabled="loading"
          @click="loadCameras"
        >
          {{ loading ? 'Refreshing…' : 'Refresh' }}
        </button>
      </div>

      <div v-if="loading && !cameras.length" class="p-8 text-sm text-gray-500">Loading cameras…</div>
      <div v-else-if="!cameras.length" class="p-8 text-center text-sm text-gray-500">
        No cameras configured yet.
      </div>
      <div v-else class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-800">
          <thead class="bg-gray-900/40">
            <tr class="text-left text-xs uppercase tracking-wide text-gray-500">
              <th class="px-4 py-3 font-medium">Name</th>
              <th class="px-4 py-3 font-medium">Video source</th>
              <th class="px-4 py-3 font-medium">Status</th>
              <th class="px-4 py-3 font-medium">Updated</th>
              <th class="px-4 py-3 font-medium text-right">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-800">
            <tr v-for="camera in cameras" :key="camera.camera_id" class="text-sm">
              <td class="px-4 py-3">
                <p class="font-medium text-gray-100">{{ camera.camera_name }}</p>
                <p class="mt-0.5 text-xs text-gray-500 font-mono">ID {{ camera.camera_id }}</p>
              </td>
              <td class="px-4 py-3 max-w-md">
                <p class="truncate font-mono text-xs text-gray-300" :title="camera.video_source">
                  {{ camera.video_source }}
                </p>
              </td>
              <td class="px-4 py-3">
                <span
                  class="inline-flex rounded-md border px-2 py-0.5 text-xs font-medium"
                  :class="camera.is_active ? 'border-emerald-800/50 bg-emerald-950/30 text-emerald-200' : 'border-gray-700 bg-gray-900/60 text-gray-400'"
                >
                  {{ camera.is_active ? 'Active' : 'Inactive' }}
                </span>
              </td>
              <td class="px-4 py-3 text-xs text-gray-500">
                {{ formatDate(camera.updated_at || camera.created_at) }}
              </td>
              <td class="px-4 py-3">
                <div class="flex justify-end gap-2">
                  <button
                    type="button"
                    class="rounded border border-gray-700 px-3 py-1.5 text-xs text-gray-300 hover:bg-gray-800"
                    @click="openEdit(camera)"
                  >
                    Edit
                  </button>
                  <button
                    type="button"
                    class="rounded border border-red-900/50 px-3 py-1.5 text-xs text-red-300 hover:bg-red-950/30"
                    @click="onDelete(camera)"
                  >
                    Delete
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>

    <Teleport to="body">
      <div
        v-if="showModal"
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
        @click.self="closeModal"
      >
        <div class="w-full max-w-md rounded-xl border border-gray-700 bg-[#12181f] p-5 shadow-xl">
          <h3 class="text-lg font-semibold text-gray-100">
            {{ modalMode === 'create' ? 'Add camera' : 'Edit camera' }}
          </h3>

          <div class="mt-4 space-y-4">
            <div>
              <label class="text-xs text-gray-500 uppercase">Camera name <span class="text-red-400">*</span></label>
              <input
                v-model="form.camera_name"
                type="text"
                class="mt-1 w-full rounded-lg border border-gray-700 bg-gray-900/80 px-3 py-2 text-sm text-gray-200"
                placeholder="cam1"
              >
            </div>
            <div>
              <label class="text-xs text-gray-500 uppercase">Video source <span class="text-red-400">*</span></label>
              <input
                v-model="form.video_source"
                type="text"
                class="mt-1 w-full rounded-lg border border-gray-700 bg-gray-900/80 px-3 py-2 text-sm text-gray-200"
                placeholder="rtsp://camera.local/stream or /dev/video0"
              >
            </div>
            <label class="inline-flex items-center gap-2 text-sm text-gray-300">
              <input v-model="form.is_active" type="checkbox" class="rounded border-gray-600">
              Active
            </label>
          </div>

          <div class="mt-6 flex justify-end gap-2">
            <button
              type="button"
              class="rounded-lg border border-gray-600 px-4 py-2 text-sm text-gray-300"
              :disabled="saving"
              @click="closeModal"
            >
              Cancel
            </button>
            <button
              type="button"
              class="rounded-lg bg-teal-600 px-4 py-2 text-sm font-medium text-white hover:bg-teal-500 disabled:opacity-50"
              :disabled="saving || !form.camera_name.trim() || !form.video_source.trim()"
              @click="submit"
            >
              {{ saving ? 'Saving…' : 'Save' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import type { CameraConfig } from '~/types/security'

definePageMeta({
  layout: 'security',
  middleware: 'auth',
})

const api = useApi()

const cameras = ref<CameraConfig[]>([])
const loading = ref(false)
const saving = ref(false)
const pageError = ref<string | null>(null)

const showModal = ref(false)
const modalMode = ref<'create' | 'edit'>('create')
const editingId = ref<number | null>(null)

const form = reactive({
  camera_name: '',
  video_source: '',
  is_active: true,
})

onMounted(() => {
  void loadCameras()
})

function resetForm() {
  form.camera_name = ''
  form.video_source = ''
  form.is_active = true
  editingId.value = null
}

function openCreate() {
  resetForm()
  modalMode.value = 'create'
  showModal.value = true
}

function openEdit(camera: CameraConfig) {
  modalMode.value = 'edit'
  editingId.value = camera.camera_id
  form.camera_name = camera.camera_name
  form.video_source = camera.video_source
  form.is_active = camera.is_active
  showModal.value = true
}

function closeModal() {
  if (saving.value) return
  showModal.value = false
}

function formatDate(iso?: string) {
  if (!iso) return '—'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' })
}

async function loadCameras() {
  loading.value = true
  pageError.value = null
  try {
    const res = await api.getCameraConfigs()
    cameras.value = res.cameras ?? []
  }
  catch (e) {
    pageError.value = e instanceof Error ? e.message : 'Failed to load cameras'
    cameras.value = []
  }
  finally {
    loading.value = false
  }
}

async function submit() {
  saving.value = true
  pageError.value = null
  const body = {
    camera_name: form.camera_name.trim(),
    video_source: form.video_source.trim(),
    is_active: form.is_active,
  }
  try {
    if (modalMode.value === 'create') {
      await api.createCamera(body)
    }
    else if (editingId.value != null) {
      await api.updateCamera(editingId.value, body)
    }
    showModal.value = false
    await loadCameras()
  }
  catch (e) {
    pageError.value = e instanceof Error ? e.message : 'Failed to save camera'
  }
  finally {
    saving.value = false
  }
}

async function onDelete(camera: CameraConfig) {
  if (!confirm(`Delete camera "${camera.camera_name}"?`)) return
  pageError.value = null
  try {
    await api.deleteCamera(camera.camera_id)
    await loadCameras()
  }
  catch (e) {
    pageError.value = e instanceof Error ? e.message : 'Failed to delete camera'
  }
}
</script>
