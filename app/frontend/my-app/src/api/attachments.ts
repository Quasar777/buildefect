// src/api/attachments.ts
import api from './axios';

export interface Attachment {
  id: number;
  defect_id?: number;
  url: string; // может быть "internal/uploads/..."
  filename?: string;
  content_type?: string;
  size?: number;
}

export const getAttachmentsByDefect = async (defectId: number): Promise<Attachment[]> => {
  const { data } = await api.get(`/defects/${defectId}/attachments`);
  return data;
};
